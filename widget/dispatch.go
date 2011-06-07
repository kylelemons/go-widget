package widget

import (
	"http"
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.HandleFunc("/widget/list", myWidgets)
	http.HandleFunc("/widget/add", addWidget)
	http.HandleFunc("/widget/show/", showWidget)
	http.HandleFunc("/widget/update/", updateWidget)

	http.HandleFunc("/hook/", fourOhFour)
	http.HandleFunc("/hook/compile/", hookCompile)
	http.HandleFunc("/hook/commit/", hookCommit)
	http.HandleFunc("/hook/plusone/", hookRating)
	http.HandleFunc("/hook/wontbuild/", hookBroken)

	http.HandleFunc("/task/", fourOhFour)
	http.HandleFunc("/task/summary", taskSummary)
	http.HandleFunc("/task/upgrade", taskUpgrade)
	http.HandleFunc("/task/wipe/datastore", taskWipe)

	// TODO(kevlar): Remove things that don't build with release
	// http://go.googlecode.com/hg/.hgtags | grep release\. | sort -n | tail -n 1
}

func root(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/widget/list", http.StatusFound)
}

func fourOhFour(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Found: " + r.URL.Path, http.StatusNotFound)
}
