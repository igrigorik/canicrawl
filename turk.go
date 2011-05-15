package main

import (
  "os"
  "fmt"
  // "log"
  "http"
  "io/ioutil"
  "github.com/temoto/robotstxt.go"
)

func handler(w http.ResponseWriter, r *http.Request) {
  error := func(msg string) {
    w.WriteHeader(400)
    fmt.Fprintf(w, msg)
  }

  query, _ := http.ParseQuery(r.URL.RawQuery)

  url, ok_url := query["url"]
  bot, ok_bot := query["agent"]
  if !ok_url || !ok_bot {
    error("Required parameters: url, agent")
    return
  }

  fmt.Printf("Handling request for %s, agent: %s\n", url, bot)

  uri, err := http.ParseURL(url[0])
  if err != nil {
    error("Invalid URL: " + err.String())
    return
  }

  robotsUri := "http://" + uri.Host + "/robots.txt"
  resp, _, err := http.Get(robotsUri)
  if err != nil {
    error("Cannot fetch robots.txt: " + err.String())
    return
  }

  fmt.Printf("\tfetched robots.txt: %s, status code: %s \n", robotsUri, resp.StatusCode)

  body, _ := ioutil.ReadAll(resp.Body)
  robots, err := robotstxt.FromResponse(resp.StatusCode, string(body), true)
  if err != nil {
    error("Cannot parse robots file")
    return
  }

  allow, err := robots.TestAgent(uri.Path, bot[0])
  if err != nil {
    error("Error evaluating agent")
    return
  }

  if !allow {
    // return 400 if robots.txt does not allow fetching this resource
    w.WriteHeader(400)
  }

}

func main() {
  // log.Printf("Starting Turk server")

  http.HandleFunc("/", handler)
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    fmt.Fprintf(os.Stderr, "ListenAndServe error: %s\n", err.String())
    os.Exit(1)
  }
}
