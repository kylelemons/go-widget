package widget

import (
	"fmt"
	"http"
	"strings"

	"appengine"
)

func hookCompile(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	_ = ctx

	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/", -1)
	if len(path) != 4 {
		http.Error(w, "ID/Hash required", http.StatusBadRequest)
		return
	}

	widgethash, archhash := path[2], path[3]
	if len(widgethash) != 32 {
		http.Error(w, "Invalid widget id: " + widgethash, http.StatusBadRequest)
		return
	}
	if len(archhash) != 32 {
		http.Error(w, "Invalid hash: " + archhash, http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr
	if ip == "" {
		ip = "devel"
	}

	if _, err := LoadWidget(ctx, widgethash); err != nil {
		http.Error(w, "Unknown widget id: " + widgethash, http.StatusBadRequest)
		return
	}

	arch := NewArchitecture(ctx, widgethash, archhash)
	if err := arch.Commit(); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	comp := NewCompilation(ctx, widgethash, ip)
	if err := comp.Commit(); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK")
}

func hookCommit(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	_ = ctx

	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/", -1)
	if len(path) != 3 {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	widgethash := path[2]
	if len(widgethash) != 32 {
		http.Error(w, "Invalid widget id: " + widgethash, http.StatusBadRequest)
		return
	}

	if _, err := LoadWidget(ctx, widgethash); err != nil {
		http.Error(w, "Unknown widget id: " + widgethash, http.StatusBadRequest)
		return
	}

	comm := NewCommit(ctx, widgethash)
	if err := comm.Commit(); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK")
}
