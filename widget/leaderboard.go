package widget

import (
	"fmt"
	"http"
	"os"
	"template"

	"appengine"
)

var leaderBoardTemplate = ``+
`<html>
<head>
	<title>Project Leader Board</title>
{CSS}
</head>
<body>
{Header}
<h1>Project Leader Board</h1>
<table class='leaderBoard'>
<thead>
	<tr>
		<th></th>
		<th>Project</th>
		<th>Score</th>
		<th>Rating</th>
		<th></th>
		<th></th>
	</th>
</thead>
<tbody>
{.repeated section Widget}
	<tr>
		<th class='topWidget left'></th>
		<th><a href="{HomeURL|html}">{Name}</a></th>
		<td class='right'>{CachedScore}/5</td>
		<td class='right'>{CachedRating} (<a href="/hook/plusone/{ID}">+</a>)</td>
		<td><a href="{SourceURL|html}">Source</a></td>
		<td><a href="{BugURL|html}">Report a Bug</a></td>
	</tr>
{.end}
</tbody>
</table>
</body>
</html>
`

type leaderBoardData struct {
	CSS string
	Header string
	Widget []*Widget
}

func leaderBoard(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	ctx := appengine.NewContext(r)

	page, err := template.Parse(leaderBoardTemplate, nil)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	data := leaderBoardData{
		CSS: commonCSS(),
		Header: header(ctx),
	}

	data.Widget, err = LoadTopWidgets(ctx)
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
