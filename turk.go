package main

import (
  "os"
  "gob"
  "log"
  "http"
  "flag"
  "bytes"
  "io/ioutil"
  "github.com/temoto/robotstxt.go"
  "github.com/kklis/gomemcache"
)

var (
  host = flag.String("host", "localhost:8080", "listening port and hostname that will appear in the urls")
  help = flag.Bool("h", false, "show this help")
  logf  = flag.String("log", "turk.log", "name and location of logfile")
  pool *Pool
)

type Robot struct {
  Resp string
  Code int
}

type Pool struct {
  newFn func() (*memcache.Memcache, os.Error)
  conns chan *memcache.Memcache
}

type pooledConnection struct {
  conn *memcache.Memcache
  pool *Pool
}

func NewPool(newFn func() (*memcache.Memcache, os.Error), maxIdle int) *Pool {
  return &Pool{newFn: newFn, conns: make(chan *memcache.Memcache, maxIdle)}
}

func NewDialPool(addr string, port int, maxIdle int) *Pool {
  connect := func(addr string, port int) *memcache.Memcache {
    conn, err := memcache.Connect(addr, port)
    if err != nil {
      log.Println("Cannot connect to memcache, error: ", err.String())
      os.Exit(1)
    }
    return conn
  }

  return NewPool(func() (*memcache.Memcache, os.Error) { return connect(addr, port), nil }, maxIdle)
}

func (p *Pool) Get() (*pooledConnection, os.Error) {
  var c *memcache.Memcache
  select {
  case c = <-p.conns:
  default:
    var err os.Error
    c, err = p.newFn()
    if err != nil {
      return nil, err
    }
  }
  return &pooledConnection{c, p}, nil
}

func (c *pooledConnection) Close() os.Error {
  if c.conn == nil {
    return nil
  }

  select {
  case c.pool.conns <- c.conn:
  default:
    c.conn.Close()
  }
  c.conn = nil
  return nil
}

// HTTP handler

func testAgent(path string, agent string, robots *robotstxt.RobotsData, w http.ResponseWriter) {
  allow, err := robots.TestAgent(path, agent)
  if (err != nil) || !allow {
    w.WriteHeader(400)
  }
}

func handler(w http.ResponseWriter, r *http.Request) {
  error := func(msg string) {
    w.WriteHeader(400)
    w.Write([]byte(msg))
  }

  query, _ := http.ParseQuery(r.URL.RawQuery)

  url, ok_url := query["url"]
  bot, ok_bot := query["agent"]
  if !ok_url || !ok_bot {
    error("Required parameters: url, agent")
    return
  }

  log.Printf("Handling request for %s, agent: %s\n", url, bot)

  uri, err := http.ParseURL(url[0])
  if err != nil {
    error("Invalid URL: " + err.String())
    return
  }

  robotsUri := "http://" + uri.Host + "/robots.txt"

  cache, err := pool.Get()
  if err != nil {
    error("Cannot acquire memcached connection: " + err.String())
    return
  }
  defer cache.Close()

  data, _, err := cache.conn.Get(robotsUri)
  if data != nil {
    log.Println("Found robots.txt data in cache for: ", robotsUri)
    decoder := gob.NewDecoder(bytes.NewBuffer(data))

    var robot Robot
    if err = decoder.Decode(&robot); err != nil {
      log.Fatal("decode error:", err)
    }

    parsed, err := robotstxt.FromResponse(robot.Code, robot.Resp, true)
    if err != nil {
      error("Cannot parse robots file: " + err.String())
      return
    }

    testAgent(uri.Path, bot[0], parsed, w)
    return
  }

  resp, _, err := http.Get(robotsUri)
  if err != nil {
    error("Cannot fetch robots.txt: " + err.String())
    return
  }

  log.Printf("\tfetched robots.txt: %s, status code: %s \n", robotsUri, resp.StatusCode)

  body, _ := ioutil.ReadAll(resp.Body)
  robots, err := robotstxt.FromResponse(resp.StatusCode, string(body), true)
  if err != nil {
    error("Cannot parse robots file: " + err.String())
    return
  }

  // cache for future requests
  var robotsGob bytes.Buffer
  encoder := gob.NewEncoder(&robotsGob)
  if err = encoder.Encode(Robot{Code: resp.StatusCode, Resp: string(body)}); err != nil {
    error("Cannot gob robots file: " + err.String())
    return
  }

  if err = cache.conn.Set(robotsUri, []uint8(robotsGob.String()), 0, 60*60*24*30); err != nil {
    error("Cannot store robots gob in memcached: " + err.String())
    return
  }

  testAgent(uri.Path, bot[0], robots, w)
}

func usage() {
  println("turk usage:")
  flag.PrintDefaults()
  os.Exit(2)
}

func main() {
  flag.Parse()
  if *help {
    usage()
  }

  log.Println("Starting Turk server on " + *host)

  out, _ := os.OpenFile(*logf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
  log.SetOutput(out)

  pool = NewDialPool("127.0.0.1", 11211, 10)

  http.HandleFunc("/", handler)
  err := http.ListenAndServe(*host, nil)
  if err != nil {
    log.Println("ListenAndServe error: %s\n", err.String())
    os.Exit(1)
  }
}
