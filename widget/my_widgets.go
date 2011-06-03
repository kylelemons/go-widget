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
</pre>
{.repeated section Widget}
<pre>
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

<style type='text/css'>
.gowidget, .gowidget *
{.meta-left}
	font-weight: normal;
	color: {Colors.Main.Text};
	background: {Colors.Main.Background};
	border: 1px solid {Colors.Main.Border};
	border-collapse: collapse;
	border-spacing: 0;
	padding: 0;
	margin: 0;
{.meta-right}

.gowidget
{.meta-left}
	margin: 5px;
{.meta-right}

.gowidget th, .gowidget td
{.meta-left}
	padding: 2px 8px;
	text-align: center;
{.meta-right}

.gowidget th
{.meta-left}
	font-weight: bold;
{.meta-right}

.gowidget thead th
{.meta-left}
	color: {Colors.Good.Text};
	background: {Colors.Good.Background};
	border: 1px solid {Colors.Good.Border};
{.meta-right}

.gowidget tbody tr th:first-child
{.meta-left}
	text-align: right;
{.meta-right}
</style>
<table class='gowidget'>
	<thead>
		<tr>
			<th colspan="3">{Name} - 5/5</td>
		</tr>
	</thead>
	<tbody>
		<tr>
			<th></th>
			<th>Build</th>
			<th>Commit</th>
		</tr>
		<tr>
			<th>Weekly:</th>
			<!--InstallWeek-->
			<td>{CompileWeek}</td>
			<td>{CheckinWeek}</td>
		</tr>
		<tr>
			<th>Total:</th>
			<!--InstallTotal-->
			<td>{CompileTotal}</td>
			<td>{CheckinTotal}</td>
		</tr>
		<tr>
			<th>Last:</th>
			<!--InstallElapsed-->
			<td>{CompileElapsed}</td>
			<td>{CheckinElapsed}</td>
		</tr>
	</tbody>
</table>

{.end}
<form method="post" action="/widget/add">
	<input name="name" />
	<input type="submit" value="add" />
</form>
</body>
</html>
`

type Pallete struct {
	Text, Text2, Border, Light, Background string
}

type colorScheme struct {
	Main, Good, Warn, Bad Pallete
}

var scheme = colorScheme{
	Main: Pallete{"#2E733E", "#30663C", "#1E602D", "#8BDD9D", "#B3DDBC"},
	Good: Pallete{"#26585C", "#274E52", "#19494E", "#89D1D8", "#AFD4D8"},
	Warn: Pallete{"#98683D", "#866140", "#805128", "#E6B991", "#E6CFBA"},
	Bad:  Pallete{"#98463D", "#864640", "#803028", "#E69991", "#E6BEBA"},
}

type myWidgetData struct {
	Widget []*Widget
	Colors colorScheme
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
		Colors: scheme,
	}

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
