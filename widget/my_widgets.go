package widget

import (
	"fmt"
	"http"
	"os"
	"regexp"
	"strings"
	"template"

	"appengine"
)

var myWidgetTemplate = ``+
`<html>
<head>
	<title>My Widgets</title>
{CSS}
</head>
<body>
{Header}
<h1>My Widgets</h1>
{.repeated section Widget}
<hr/>
<h2>{Name}</h2>
<h3>Embed:</h3>
<pre>
&lt;script language="javascript" type="text/javascript"
	source="http://go-widget.appspot.com/widget/show/{ID}/widget.js">
&lt;/script>
&lt;noscript>
&lt;a href="http://go-widget.appspot.com/widget/show/{ID}">{Name}&lt;/a>
&lt;/noscript>
</pre>
<h3>Rating:</h3>
<ol>
<li>Rated at least +5 (Current: {Rating})</li>
<li>At least 50 compiles (Current: {CompileTotal})</li>
<li>At least 5 compile at HEAD (Current: {CompileCheckin})`+/*TODO(kevlar): since Go release*/`
<li>No more than 1 "won't build" at HEAD (Current: {Broken})</li>
<li>Set Home, Source, and Bug Report URLs</li>
</ol>
<h3>URLs</h3>
<form method="post" action="/widget/update/{ID}">
<table>
<tr><td>Home:</td>
<td><input name="home" type="text" size="100" value="{HomeURL|html}"/></td></tr>
<tr><td>Source:</td>
<td><input name="source" type="text" size="100" value="{SourceURL|html}"/></td></tr>
<tr><td>Create Bug:</td>
<td><input name="bug" type="text" size="100" value="{BugURL|html}"/></td></tr>
</table>
<input type="submit" value="Update"/>
</form>
<h3>Hooks</h3>
Commit Hook URL: <pre>http://go-widget.appspot.com/hook/commit/{ID}</pre>
Makefile:
<pre>
--- %< ---
success :
	@curl -s "http://go-widget.appspot.com/hook/compile/{ID}" >/dev/null
--- %< ---
</pre>

<!--

Primary Color: (green)
28DE5F	40A65F	0D9035	5CEE88	83EEA4
Secondary Color A: (blue)
2AAFCE	3F899B	0E6F86	5DCDE7	82D4E7
Secondary Color B: (orange)
FF962F	BF844A	A65A0F	FFB063	FFC58C
Complementary Color: (red)
FF4B2F	BF5A4A	A6240F	FF7863	FF9C8C
(accent, saturated, dark, pastel, light)

-->

{ExecuteString}

{.end}
<form method="post" action="/widget/add">
	<input name="name" />
	<input type="submit" value="add" />
</form>
</body>
</html>
`

type myWidgetData struct {
	CSS string
	Header string
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

	data := myWidgetData{
		CSS: commonCSS(),
		Header: header(ctx),
	}

	data.Widget, err = LoadWidgets(ctx)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if len(r.FormValue("ids_only")) > 0 {
		w.Header().Set("Content-Type", "text/plain")
		for _, widget := range data.Widget {
			fmt.Fprintln(w, widget.ID)
		}
		return
	}

	page.Execute(w, data)
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

	if len(name) == 0 {
		http.Error(w, "Invalid/no name provided", http.StatusInternalServerError)
		return
	}

	widget := NewWidget(ctx, name)

	err = widget.Commit()
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/widget/list", http.StatusFound)
}

func showWidget(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	_ = ctx

	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/", -1)
	if cnt := len(path); cnt < 3 || cnt > 4 {
		http.Error(w, "ID required, widget.js optional", http.StatusBadRequest)
		return
	}
	nojs := len(path) == 3

	widgethash := path[2]
	if len(widgethash) != 32 {
		http.Error(w, "Invalid widget id: " + widgethash, http.StatusBadRequest)
		return
	}

	widget, err := LoadWidget(ctx, widgethash)
	if err != nil {
		http.Error(w, "Unknown widget id: " + widgethash, http.StatusBadRequest)
		return
	}

	if nojs {
		fmt.Fprintf(w, "<html><head><title>"+widget.Name+"</title></head><body>\n")
		fmt.Fprintf(w, "<script language='javascript' type='text/javascript'>\n")
	} else {
		w.Header().Set("Content-Type", "text/javascript")
	}
	full := widget.ExecuteString()
	full = strings.Replace(full, "</", "<'+'/'+'", -1)
	for lineno, line := range strings.Split(full, "\n", -1) {
		fmt.Fprintf(w, "/*%3d*/ document.write('%s');\n", lineno+1, line)
	}
	if nojs {
		fmt.Fprintf(w, "</script></body></html>")
	}
}

func updateWidget(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	ctx := appengine.NewContext(r)

	fix := func(formname string) string {
		raw := r.FormValue(formname)
		if strings.Contains(raw, "'") {
			ctx.Logf("Hacking attempt: %#q", raw)
			return ""
		}
		url, err := http.ParseURLReference(raw)
		if err != nil {
			ctx.Logf("Bad URL: %#q", raw)
			return ""
		}
		if scheme := strings.ToLower(url.Scheme); scheme != "http" && scheme != "https" {
			ctx.Logf("Bad scheme: %%q", scheme)
			return ""
		}
		return url.String()
	}

	bug := fix("bug")
	home := fix("home")
	source := fix("source")

	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/", -1)
	if cnt := len(path); cnt != 3 {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	widgethash := path[2]
	if len(widgethash) != 32 {
		http.Error(w, "Invalid widget id: " + widgethash, http.StatusBadRequest)
		return
	}

	widget, err := LoadWidget(ctx, widgethash)
	if err != nil {
		http.Error(w, "Unknown widget id: " + widgethash, http.StatusBadRequest)
		return
	}

	widget.BugURL = bug
	widget.HomeURL = home
	widget.SourceURL = source

	err = widget.Commit()
	if err != nil {
		http.Error(w, "Error comitting: " + err.String(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/widget/list", http.StatusFound)
}
