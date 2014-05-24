package main

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"net/http"
)

type sessionKeyType int

const sessionKey sessionKeyType = 0

// Sessions is a Negroni middleware that handles session lifecycles for requests
type Sessions struct {
	Name         string
	Store        *sessions.CookieStore
	ClearContext bool
}

// New Sessions returns a new instance of Sessions for a given session name and cookie store
// The "clearContext" param controls whether context.Clear(r) is run during the unwind
// to clear context values for the request.
// If you're using gorilla/mux, set as false
func NewSessions(name string, store *sessions.CookieStore, clearContext bool) *Sessions {
	return &Sessions{name, store, clearContext}
}

// GetSession returns the session instance for the request from it's context
func GetSession(r *http.Request) *sessions.Session {
	if v := context.Get(r, sessionKey); v != nil {
		return v.(*sessions.Session)
	}
	return nil
}

// ServeHTTP retrieves or creates a session for the request, stores it in the request context,
// and then saves the session as the middleware stack unwinds.
func (s *Sessions) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	session, _ := s.Store.Get(r, s.Name)
	context.Set(r, sessionKey, session)

	// Using defer so the cleanup is always run, even if a panic occurs
	defer func() {
		session.Save(r, rw)
		if s.ClearContext {
			context.Clear(r)
		}
	}()

	next(rw, r)
}
