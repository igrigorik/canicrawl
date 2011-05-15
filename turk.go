package main

import (
  "os"
  "fmt"
  "http"
  "github.com/temoto/robotstxt.go"
)

func handler(w http.ResponseWriter, r *http.Request) {
  robotstxt.FromString("User-agent: *\nDisallow:", true)

  fmt.Printf("Handling request for %s\n", r.URL)

  fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
  fmt.Println("Starting Turk server")

  http.HandleFunc("/", handler)
  err := http.ListenAndServe(":80", nil)
  if err != nil {
    fmt.Fprintf(os.Stderr, "ListenAndServe error: %s\n", err.String())
    os.Exit(1)
  }
}
