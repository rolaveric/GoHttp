package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/codegangsta/martini"
)

// Type which, through reflection, Martini uses to dependency inject the user
// found by the Authentication middleware
type User interface{}

func main() {
	// Starting a standard martini server,
	// with logging and panic recovery middleware built in
	m := martini.Classic()

	// Register middleware which authenticates the client
	m.Use(Authentication)

	// Register a route which does something with the authenticated user,
	// but doesn't care for security reasons (ie. No authorization check)
	m.Get("/", func(u User) string {
		return fmt.Sprintf("Hello %s!", u)
	})

	// Register a route which requires the authenticated user to have specific authorization
	m.Get("/secret", Authorization("secret access"), func(u User) string {
		return fmt.Sprintf("Hello Secret Agent %s!", u)
	})

	// Start the server
	m.Run()
}

// Middleware for Martini which tries to identify the request user by the Authorization header.
// If no such header exists, it assumes it's a guest.
// If the header exists but can't be decoded, it returns a 401 status.
// It uses Martini's dependency injection system to register the user for later consumption.
func Authentication(c martini.Context, req *http.Request, res http.ResponseWriter) {
	// Get the Authorization header
	a := req.Header.Get("Authorization")
	if a == "" {
		// No header, user must be guest
		c.MapTo("guest", (*User)(nil))
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
