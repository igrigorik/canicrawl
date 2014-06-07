package turk

import (
  "appengine"
  "appengine/urlfetch"
  "github.com/temoto/robotstxt.go"
  "net/http"
  "net/url"
  "encoding/json"
)

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

  req_agent := r.Header.Get("User-Agent")
  req_url, ok_url := query["url"]

  if !ok_url {
    error("required parameters: url")
    return
  }

  c.Infof("Handling request for %s, agent: %s\n", req_url, req_agent)

  parsed_url, err := url.Parse(req_url[0])
  if err != nil {
    error("Invalid URL: " + err.Error())
    return
  }

  robotsUrl := "http://" + parsed_url.Host + "/robots.txt"

  client := urlfetch.Client(c)
  resp, err := client.Get(robotsUrl)
  if err != nil {
      error("cannot fetch robots.txt: " + err.Error())
      return
  }

  c.Infof("Fetched robots.txt: %s, status code: %s \n", robotsUrl, resp.StatusCode)
  defer resp.Body.Close()

  robots, err := robotstxt.FromResponse(resp)
  if err != nil {
    error("cannot parse robots file: " + err.Error())
    return
  }

  allow := robots.TestAgent(parsed_url.Path, req_agent)
  if !allow {
    reply := map[string] string {
        "status":  "disallowed",
    }
    resp, _ := json.Marshal(reply)

    w.WriteHeader(400)
    w.Write(resp)
    return
  }

  w.Header().Set("Location", req_url[0])
  w.WriteHeader(http.StatusFound)
}

func init() {
    http.Handle("/", http.FileServer(http.Dir("static")))
    http.HandleFunc("/check", handler)
}
