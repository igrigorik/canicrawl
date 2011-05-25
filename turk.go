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

type Robot struct {
  Resp string
  Code int
}

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

  // this is hardly efficient, but its a first pass..
  // open a new memcache connection on each request and
  // fetch the serialized robots rules
  cache, err := memcache.Connect("127.0.0.1", 11211)
  if err != nil {
    error("Cannot connect to memcached: " + err.String())
    return
  }

  data, _, err := cache.Get(robotsUri)
  if data != nil {
    log.Println("Found robots.txt data in cache for: ", robotsUri)
    decoder := gob.NewDecoder(bytes.NewBuffer(data))

    var robot Robot
    err = decoder.Decode(&robot)
    if err != nil {
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
  err = encoder.Encode(Robot{Code: resp.StatusCode, Resp: string(body)})
  if err != nil {
    error("Cannot gob robots file: " + err.String())
    return
  }

  err = cache.Set(robotsUri, []uint8(robotsGob.String()), 0, 60*60*24*30)
  if err != nil {
    error("Cannot store robots gob in memcached: " + err.String())
    return
  }

  testAgent(uri.Path, bot[0], robots, w)
}

var (
  host = flag.String("host", "localhost:8080", "listening port and hostname that will appear in the urls")
  help = flag.Bool("h", false, "show this help")
)

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

  http.HandleFunc("/", handler)
  err := http.ListenAndServe(*host, nil)
  if err != nil {
    log.Println("ListenAndServe error: %s\n", err.String())
    os.Exit(1)
  }
}
