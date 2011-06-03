package widget

import (
	"http"
	"strings"
	"appengine"
	"appengine/user"
)

func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, "/")
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		// Temporary band-aid for development
		url, err = http.URLUnescape(url)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		url = strings.Replace(url, "http://localhost:8080", "", 1)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u != nil {
		url, err := user.LogoutURL(c, "/")
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		// Temporary band-aid for development
		url, err = http.URLUnescape(url)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		url = strings.Replace(url, "http://localhost:8080", "", 1)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
