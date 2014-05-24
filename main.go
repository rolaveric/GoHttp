package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"net/http"
)

func main() {
	// Starting a standard negroni stack with logging and panic recovery middleware.
	n := negroni.Classic()

	// Register middleware for storing session details in cookies
	// NOTE: In a real world example, you do NOT want to store all your session
	// data in cookies.  Just an ID or Token, and store everything else in a DB.
	// Other session store implementations at: https://github.com/gorilla/sessions
	store := sessions.NewCookieStore([]byte("secret123"))
	n.Use(NewSessions("github.com/rolaveric/GoHttp", store, false))

	// Register middleware which authenticates the client
	n.Use(NewAuthentication())

	// Starting a Gorilla mux instance for routing with
	r := mux.NewRouter()

	// Register a route which does something with the authenticated user,
	// but doesn't care for security reasons (ie. No authorization check)
	r.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		u := GetUser(req)
		fmt.Fprintf(rw, "Hello %s!", u)
	}).Methods("GET")

	// Register a route which requires the client to use Basic Authentication.
	// If they don't, respond with 401 status and a WWW-Authenticate header
	r.HandleFunc("/login", func(rw http.ResponseWriter, req *http.Request) {
		u := GetUser(req)
		// If they're already logged in, redirect to "/"
		if u != AnonUser {
			http.Redirect(rw, req, "/", 302)
			return
		}

		// Otherwise, prompt for login
		rw.Header().Set("WWW-Authenticate", "Basic realm=\"Authentication Required\"")
		http.Error(rw, "Basic Authentication Required", http.StatusUnauthorized)
	}).Methods("GET")

	// Register a route for clearing the session
	r.HandleFunc("/logout", func(rw http.ResponseWriter, req *http.Request) {
		s := GetSession(req)
		s.Options.MaxAge = -1
		fmt.Fprintf(rw, "Logged out")
		//http.Redirect(rw, req, "/", 302)
	}).Methods("GET")

	// Register a route which requires the authenticated user to have specific authorization
	/*n.HandleFunc("/secret", Authorization("secret access"), func(u User) string {
		return fmt.Sprintf("Hello Secret Agent %s!", u)
	}).Methods("GET")*/

	// Telling negroni to use the Gorilla mux as it's handler
	n.UseHandler(r)

	// Start the server
	n.Run(":3000")
}

/*


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
}*/
