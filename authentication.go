package main

// Middleware for Negroni which tries to identify the request user by the Authorization header.
// If no such header exists, it checks the session cookie.  If that also fails, it assumes the user is a guest.
// If the header exists but can't be decoded, it returns a 401 status.
// It uses gorilla/context to retrieve the session and store the user data.

import (
	"encoding/base64"
	"github.com/gorilla/context"
	"net/http"
	"strings"
)

type userKeyType int

const userKey userKeyType = 0

// Type which for the user found by the Authentication middleware
type User struct {
	Name string
}

func (u *User) String() string {
	return u.Name
}

type Authentication struct{}

func NewAuthentication() *Authentication {
	return &Authentication{}
}

// GetUser returns the user instance for the request from it's context
func GetUser(r *http.Request) *User {
	if v := context.Get(r, userKey); v != nil {
		return v.(*User)
	}
	return nil
}

var AnonUser *User = &User{"guest"}

func (mw *Authentication) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	session := GetSession(r)

	// Get the Authorization header
	a := r.Header.Get("Authorization")
	if a == "" {
		// No header - Check for a session
		u := session.Values["user"]
		if u == nil {
			// No session either, set the user as a guest
			context.Set(r, userKey, AnonUser)
		} else {
			// Register the user from the session
			context.Set(r, userKey, &User{u.(string)})
		}
		next(rw, r)
		return
	}

	// Remove the 'Basic ' prefix
	a = strings.TrimPrefix(a, "Basic ")

	// Decode the Authorization header
	data, err := base64.StdEncoding.DecodeString(a)
	if err != nil {
		// Bad Authorization value - return 401 Unauthorized
		http.Error(rw, "Could not decode Authorization header", http.StatusUnauthorized)
		// Note: If the response gets written to, Negroni doesn't call any subsequent handlers
		return
	}

	// Split out the username and password
	s := strings.Split(string(data), ":")
	if len(s) < 2 {
		// Bad Authorization value - return 401 Unauthorized
		http.Error(rw, "Authorization header requires a username and password", http.StatusUnauthorized)
		// Note: If the response gets written to, Negroni doesn't call any subsequent handlers
		return
	}

	// At this point, you would normally lookup a user database
	// Since this is an example - we're just accepting the username as is
	context.Set(r, userKey, &User{s[0]})

	// Set the user in the session
	session.Values["user"] = s[0]

	next(rw, r)
}
