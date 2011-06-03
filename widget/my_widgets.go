package widget

import (
	"fmt"
	"http"
	"os"
	"regexp"
	"template"

	"appengine"
)

var myWidgetTemplate = ``+
`<html>
<head>
	<title>My Widgets</title>
</head>
<body>
<pre>
=== My Widgets ===
{.repeated section Widget}
Widget: {Name}
	Repository Checkins:
		Total:     {CheckinTotal}
		This Week: {CheckinWeek}
		Last:      {CheckinDate}
	Successful Compiles:
		Total:     {CompileTotal}
		This Week: {CompileWeek}
		At HEAD:   {CompileCheckin}
		Last:      {CompileDate}
	Hook URLs:
		Commit:    http://go-widget.appspot.com/hook/commit/{ID}
	Makefile:
--- %< ---
success :
	@curl -s "http://go-widget.appspot.com/hook/compile/{ID}/$$(uname -srm | md5)" >/dev/null
--- %< ---

{.end}
</pre>
<form method="post" action="/widget/add">
	<input name="name" />
	<input type="submit" value="add" />
</form>
</body>
</html>
`

type myWidgetData struct {
	Widget []*Widget
}

func myWidgets(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	ctx := appengine.NewContext(r)

	page, err := template.Parse(myWidgetTemplate, nil)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	data := myWidgetData{}

	data.Widget, err = LoadWidgets(ctx)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	page.Execute(w, data)

	fmt.Fprintf(w, "<!--\n%#v\n-->", r)
}

// TODO(kevlar): Move to init
var fixup *regexp.Regexp

func addWidget(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	if fixup == nil {
		fixup, err = regexp.Compile(`[^A-Za-z0-9\-_. ]`)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
	}

	ctx := appengine.NewContext(r)
	name := fixup.ReplaceAllString(r.FormValue("name"), "")

	widget := NewWidget(ctx, name)

	err = widget.Commit()
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/widget/list", http.StatusFound)
}
