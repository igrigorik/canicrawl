# Turk: robots.txt permission verifier

Simple (Golang) HTTP web-service to verify whether a supplied "Agent" is allowed to access the requested URL. Pass in the URL of the resource you want to fetch and the name of your agent and Turk will download, parse the robots.txt file and respond with a 200 if you can proceed, and 400 otherwise.

```
$> goinstall github.com/temoto/robotstxt.go
$> make && ./turk -host="localhost:9090"
$>
$> curl -v "http://127.0.0.1:9090/?agent=Googlebot&url=http://blogspot.com/comment.g"
   < HTTP/1.1 400 Bad Request

$> curl -v "http://127.0.0.1:9090/?agent=Googlebot&url=http://blogspot.com/"
   < HTTP/1.1 200 OK
```

Note: [blogger.com/robots.txt](http://blogger.com/robots.txt) blocks allow agents from fetching `comment.g` resource.

## Notes

Turk is an experiment with [Go](http://golang.org/). Go's http stack is "async", hence many parallel requests can be processed at the same time. Turk also has naive, unbounded in-memory cache to avoid refetching the same robots.txt data for a given host.

### License

(MIT License) - Copyright (c) 2011 Ilya Grigorik