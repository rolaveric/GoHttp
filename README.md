GoHttp
======

A generic HTTP server program for demonstrative purposes.

Starts a simple server which:
- Recovers from any panics that might occur while handling requests.
- Logs the requests that come in, and how long they take to handle.
- Authenticates the user by the Authorization header in the request, using Basic Authorization
- Hands control over to the router.

The router can take multiple handlers for the same route.
If one handler writes to the response, the router stops processing the subsequent handlers.
So in this case we have 2 routes setup:

GET /
- Says "Hello (user)", regardless of whether that user is a guest or not.

GET /secret
- Checks that the user has "secret access" (ie. Is their username 'admin')
- Says "Hello Secret Agent (user)"

GET /login
- If an Authorization header isn't included, it responds with a 401 status and a WWW-Authenticate header, prompting the browser provide a Basic authentication form.  If one is provided, it redirects to "/".  Demonstrates prompting for authentication, but impractical without using session cookies to persist the authentication.

I'm using this project to test out HTTP server practices with Go, and have public examples to point others to.
So any comments and constructive criticisms are welcomed.
