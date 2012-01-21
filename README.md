# Can I Crawl (this URL)

Hosted robots.txt permissions verifier.

## ENDPOINTS

- [`/`](http://canicrawl.appspot.com/) This page.
- [`/check`](http://canicrawl.appspot.com/check) Runs the robots.txt verification check.

## Description

Verifies if the provided URL is allowed to be crawled by your User-Agent. Pass in the destination URL and the service will download, parse and check the [robots.txt](http://www.robotstxt.org/) file for permissions. If you're allowed to continue, it will issue a **3XX** redirect, otherwise a **4XX** code is returned.

## Examples

### $ curl -v http://canicrawl.appspot.com/check?url=http://google.com/
	< HTTP/1.0 302 Found
	< Location: http://www.google.com/

### $ curl -v http://canicrawl.appspot.com/check?url=http://google.com/search
	< HTTP/1.0 400 Bad Request
	< Content-Length: 23
	{"status":"disallowed"}

### $ curl -H'User-Agent: MyCustomAgent' -v http://canicrawl.appspot.com/check?url=http://google.com/
	> User-Agent: MyCustomAgent
	< HTTP/1.0 302 Found
	< Location: http://www.google.com/

Note: [google.com/robots.txt](http://google.com/robots.txt) disallows requests to _/search_.

### License

MIT License - Copyright (c) 2011 [Ilya Grigorik](http://www.igvita.com/)
