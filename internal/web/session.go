package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const csrfField = "csrf_token"

// sessions holds the single-use bootstrap token and the live sessions, each
// mapped to its CSRF token.
type sessions struct {
	mu        sync.Mutex
	token     string
	tokenUsed bool
	byID      map[string]string
}

func newSessions(token string) *sessions {
	return &sessions{token: token, byID: map[string]string{}}
}

// exchange spends the bootstrap token and returns a new session ID. It
// reports false for a wrong, empty or already used token.
func (ss *sessions) exchange(token string) (string, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.tokenUsed || ss.token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(ss.token)) != 1 {
		return "", false
	}
	ss.tokenUsed = true
	id, csrf := randomToken(), randomToken()
	ss.byID[id] = csrf
	return id, true
}

// csrf returns the CSRF token of session id, or "" if there is no such session.
func (ss *sessions) csrf(id string) string {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.byID[id]
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error.
	return hex.EncodeToString(b)
}

// csrfToken returns the CSRF token of r's session, or "" without a session.
func (s *Server) csrfToken(r *http.Request) string {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return ""
	}
	return s.sess.csrf(c.Value)
}

// handleAuth exchanges the bootstrap token for a session cookie and redirects
// to / so the token leaves the address bar.
func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	id, ok := s.sess.exchange(r.URL.Query().Get("token"))
	if !ok {
		http.Error(w, "invalid or used token", http.StatusForbidden)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// requireSession rejects requests without a live session (401) and writes
// without the session's CSRF token (403) before calling next.
func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := s.csrfToken(r)
		if want == "" {
			deny(w, r, http.StatusUnauthorized, "session required")
			return
		}
		if isWrite(r.Method) {
			got := r.Header.Get("X-CSRF-Token")
			if got == "" {
				got = r.PostFormValue(csrfField)
			}
			if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
				deny(w, r, http.StatusForbidden, "missing or wrong csrf token")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// deny writes a permission_denied JSON error for /api routes and a plain HTML
// page otherwise.
func deny(w http.ResponseWriter, r *http.Request, status int, msg string) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, `{"code":"`+string(errcode.PermissionDenied)+`","error":"`+msg+`"}`+"\n")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, "<!doctype html><title>Snapback</title><p>"+msg+"</p>\n")
}
