package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
)

const ANON_USER string = "guest"

// Type which, through reflection, Martini uses to dependency inject the user
// found by the Authentication middleware
type User interface{}

func main() {
	// Starting a standard martini server,
	// with logging and panic recovery middleware built in
	m := martini.Classic()

	// Register middleware for storing session details in cookies
	// NOTE: In a real world example, you do NOT want to store all your session
	// data in cookies.  Just an ID or Token, and store everything else in a DB.
	// Other session store implementations at: https://github.com/gorilla/sessions
	store := sessions.NewCookieStore([]byte("secret123"))
	m.Use(sessions.Sessions("github.com/rolaveric/GoHttp", store))

	// Register middleware which authenticates the client
	m.Use(Authentication)

	// Register a route which does something with the authenticated user,
	// but doesn't care for security reasons (ie. No authorization check)
	m.Get("/", func(u User) string {
		return fmt.Sprintf("Hello %s!", u)
	})

	// Register a route which requires the client to use Basic Authentication.
	// If they don't, respond with 401 status and a WWW-Authenticate header
	m.Get("/login", func(u User, req *http.Request, res http.ResponseWriter) {
		// If they're already logged in, redirect to "/"
		if u != ANON_USER {
			http.Redirect(res, req, "/", 302)
			return
		}

		// Otherwise, prompt for login
		res.Header().Set("WWW-Authenticate", "Basic realm=\"Authentication Required\"")
		http.Error(res, "Basic Authentication Required", http.StatusUnauthorized)
	})

	// Register a route which requires the authenticated user to have specific authorization
	m.Get("/secret", Authorization("secret access"), func(u User) string {
		return fmt.Sprintf("Hello Secret Agent %s!", u)
	})

	// Start the server
	m.Run()
}

// Middleware for Martini which tries to identify the request user by the Authorization header.
// If no such header exists, it checks the session cookie.  If that also fails, it assumes the user is a guest.
// If the header exists but can't be decoded, it returns a 401 status.
// It uses Martini's dependency injection system to register the user for later consumption.
func Authentication(c martini.Context, req *http.Request, res http.ResponseWriter, session sessions.Session) {
	// Get the Authorization header
	a := req.Header.Get("Authorization")
	if a == "" {
		// No header - Check for a session
		u := session.Get("user")
		if u == nil {
			// No session either, set the user as a guest
			c.MapTo(ANON_USER, (*User)(nil))
		} else {
			// Register the user from the session
			c.MapTo(u, (*User)(nil))
		}
		return
	}

	// Remove the 'Basic ' prefix
	a = strings.TrimPrefix(a, "Basic ")

	// Decode the Authorization header
	data, err := base64.StdEncoding.DecodeString(a)
	if err != nil {
		// Bad Authorization value - return 401 Unauthorized
		http.Error(res, "Could not decode Authorization header", http.StatusUnauthorized)
		// Note: If the response gets written to, Martini doesn't call any subsequent handlers
		return
	}

	// Split out the username and password
	s := strings.Split(string(data), ":")
	if len(s) < 2 {
		// Bad Authorization value - return 401 Unauthorized
		http.Error(res, "Authorization header requires a username and password", http.StatusUnauthorized)
		// Note: If the response gets written to, Martini doesn't call any subsequent handlers
		return
	}

	// At this point, you would normally lookup a user database
	// Since this is an example - we're just accepting the username as is
	c.MapTo(s[0], (*User)(nil))

	// Set the user in the session
	session.Set("user", s[0])
}

// Returns a handler function which checks that the authenticated user for the request
// has a particular type of access.
// If they don't, a 401 status resposne is made.
// Otherwise, nothing happens - letting the next handler in the chain do it's work
func Authorization(access string) func(User, http.ResponseWriter) {
	// Returning a handler function that checks for the required access
	return func(u User, res http.ResponseWriter) {
		// For brevity, we're just checking that the username is 'admin'
		if u != "admin" {
			// If not, return a response that stops Martini calling the next handler
			http.Error(res, "Not Authorized", http.StatusUnauthorized)
		}
	}
}
