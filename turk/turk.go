package turk

import (
  "appengine"
  "appengine/urlfetch"  
  "robotstxt.go"
  "io/ioutil"
  "http"
  "json"
  "url"
)

func testAgent(path string, agent string, robots *robotstxt.RobotsData, w http.ResponseWriter) {
  allow, err := robots.TestAgent(path, agent)
  if (err != nil) || !allow {
    
    reply := map[string] string {
        "status":  "disallowed",
    }
    resp, _ := json.Marshal(reply)
    
    w.WriteHeader(400)
    w.Write(resp)      
    return
  }
  
  reply := map[string] string {
      "status":  "ok",
  }
  resp, _ := json.Marshal(reply)
    
  w.WriteHeader(200)
  w.Write(resp)  
}

func handler(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  c.Infof("Requested URL: %v", r.URL)

  error := func(msg string) {
    reply := map[string] string {
        "error":  msg,
    }
    resp, _ := json.Marshal(reply)
    
    w.WriteHeader(500)
    w.Write(resp)
  }

  query, _ := url.ParseQuery(r.URL.RawQuery)

  req_url, ok_url := query["url"]
  req_bot, ok_bot := query["agent"]
  if !ok_url || !ok_bot {
    error("required parameters: url, agent")
    return
  }

  c.Infof("Handling request for %s, agent: %s\n", req_url, req_bot)

  parsed_url, err := url.Parse(req_url[0])
  if err != nil {
    error("Invalid URL: " + err.String())
    return
  }

  robotsUrl := "http://" + parsed_url.Host + "/robots.txt"

  client := urlfetch.Client(c)
  resp, err := client.Get(robotsUrl)
  if err != nil {
      error("cannot fetch robots.txt: " + err.String())      
      return
  }
      
  c.Infof("Fetched robots.txt: %s, status code: %s \n", robotsUrl, resp.StatusCode)

  defer resp.Body.Close()
  body, _ := ioutil.ReadAll(resp.Body)
  
  robots, err := robotstxt.FromResponse(resp.StatusCode, string(body), true)
  if err != nil {
    error("cannot parse robots file: " + err.String())
    return
  }

  testAgent(parsed_url.Path, req_bot[0], robots, w)
}

func init() {
    http.HandleFunc("/", handler)
}