package widget

import (
	"fmt"
	"http"
	"strings"

	"appengine"
)

func hookCountable(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/", -1)
	if len(path) < 3 {
		http.Error(w, "/hook/{type}/{widget} - missing required path segment", http.StatusBadRequest)
		return
	}

	var uniqueKey int64
	var countable string
	switch path[1] {
	case "plusone":
		countable = "Rating"
	case "wontbuild":
		countable = "Broken"
	case "compile":
		countable = "Build"
		uniqueKey = int64(now())
	case "commit":
		countable = "Commit"
		uniqueKey = int64(now())
	default:
		http.Error(w, "/hook/{type}/{widget} - unknown type", http.StatusBadRequest)
		return
	}

	widget := path[2]
	if len(widget) != 32 {
		http.Error(w, "Invalid widget id: " + widget, http.StatusBadRequest)
		return
	}
	if _, err := LoadWidget(ctx, widget); err != nil {
		http.Error(w, "Unknown widget id: " + widget, http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr
	if ip == "" {
		ip = "devel"
	}

	keyhash := Hashf("IP=%s|Unique=%d", ip, uniqueKey)
	obj := NewCountable(ctx, countable, widget, keyhash)
	if err := obj.Commit(); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	refreshWidget(w,r,widget)

	// TODO(kevlar): Referer?
	//http.Redirect(w, r, "/widget/list", http.StatusFound)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK")
}
