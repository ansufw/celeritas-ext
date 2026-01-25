package middleware

import (
	"myapp/data"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *Middleware) CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Session.Exists(r.Context(), "userID") {
			// user is not logged in
			cookie, err := r.Cookie(m.App.AppName + "_remember_token")
			if err != nil {
				// no cookie, so on to next middleware
				next.ServeHTTP(w, r)

			} else {
				// we found a cookie, so let's check it
				key := cookie.Value
				var u data.User
				if len(key) > 0 {
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1]
					id, _ := strconv.Atoi(uid)
					if !u.CheckForRememberToken(id, hash) {
						m.deleteRememberCookie(w, r)
						m.App.Session.Put(r.Context(), "error", "you've been logged out from another device")
						next.ServeHTTP(w, r)
					} else {
						// valid hash
						user, _ := u.Get(id)
						m.App.Session.Put(r.Context(), "user_id", user.ID)
						m.App.Session.Put(r.Context(), "remember_token", hash)
						next.ServeHTTP(w, r)
					}
				} else {
					// key is zero, probably a leftover cookie (user has not closed browser)
					m.deleteRememberCookie(w, r)
					next.ServeHTTP(w, r)
				}
			}
		} else {
			// user is logged in
			next.ServeHTTP(w, r)
		}
	})
}

func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	newCookie := &http.Cookie{
		Name:     m.App.AppName + "_remember_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   m.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   m.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, newCookie)

	// log user out
	m.App.Session.Remove(r.Context(), "userID")
	m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
}
