package widget

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html"
	"io"
	"os"
	"template"

	"appengine"
	"appengine/user"
	"appengine/datastore"
)

type Widget struct {
	ctx appengine.Context
	key *datastore.Key

	Name  string
	ID    string
	Owner string

	HomeURL   string
	BugURL    string
	SourceURL string

	PlusOnes   int
	WontBuilds int

	CompileTotal   int
	CompileWeek    int
	CompileCheckin int
	CheckinTotal   int
	CheckinWeek    int

	LastCompile  datastore.Time
	LastCheckin  datastore.Time
	StatsUpdated datastore.Time
}

func (w *Widget) CompileDate() string {
	if w.LastCompile == 0 {
		return "never"
	}
	return timestr(w.LastCompile)
}

func (w *Widget) CheckinDate() string {
	if w.LastCheckin == 0 {
		return "never"
	}
	return timestr(w.LastCheckin)
}

func (w *Widget) CompileElapsed() string {
	if w.LastCompile == 0 {
		return "never"
	}
	elapsedHours := int64(now()-w.LastCompile) / 1e6 / 60 / 60
	elapsedHours, elapsedDays := elapsedHours%24, elapsedHours/24
	return fmt.Sprintf("%dd %dh", elapsedDays, elapsedHours)
}

func (w *Widget) CheckinElapsed() string {
	if w.LastCheckin == 0 {
		return "never"
	}
	elapsedHours := int64(now()-w.LastCheckin) / 1e6 / 60 / 60
	elapsedHours, elapsedDays := elapsedHours%24, elapsedHours/24
	return fmt.Sprintf("%dd %dh", elapsedDays, elapsedHours)
}

func (w *Widget) Commit() (err os.Error) {
	w.key, err = datastore.Put(w.ctx, w.key, w)
	return
}

func (w *Widget) Delete() (err os.Error) {
	err = datastore.Delete(w.ctx, w.key)
	return
}

func NewWidget(ctx appengine.Context, name string) *Widget {
	u := user.Current(ctx)

	sum := md5.New()
	fmt.Fprintf(sum, "Owner=%s|Widget=%s", u.Email, name)
	hash := fmt.Sprintf("%X", sum.Sum())

	return &Widget{
		ctx:   ctx,
		key:   datastore.NewKey("Widget", hash, 0, nil),
		Name:  name,
		ID:    hash,
		Owner: u.Email,
	}
}

func LoadWidget(ctx appengine.Context, id string) (widget *Widget, err os.Error) {
	key := datastore.NewKey("Widget", id, 0, nil)

	widget = new(Widget)
	err = datastore.Get(ctx, key, widget)
	widget.ctx = ctx
	widget.key = key

	return
}

func LoadWidgets(ctx appengine.Context) (widgets []*Widget, err os.Error) {
	u := user.Current(ctx)

	query := datastore.NewQuery("Widget")
	query.Filter("Owner =", u.Email)
	query.Order("Name")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &widgets)
	for i, w := range widgets {
		w.ctx = ctx
		w.key = k[i]
	}

	return
}

func LoadAllWidgets(ctx appengine.Context) (widgets []*Widget, err os.Error) {
	query := datastore.NewQuery("Widget")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &widgets)
	for i, w := range widgets {
		w.ctx = ctx
		w.key = k[i]
	}

	return
}

func (w *Widget) Rating() int {
	query := datastore.NewQuery("Rating")
	query.Filter("Widget =", w.key)
	cnt, _ := query.Count(w.ctx)
	return cnt
}

var widgetStatic string
var widgetStaticTemplate = ``+
`<style type="text/css">
.gowidget
{
	margin: 5px;
	width: 300px;
	border-collapse: collapse;
	border-spacing: 0;
}

.gowidget th, .gowidget td
{
	width: 150px;
	font-weight: normal;
	font-size: 10pt;
	margin: 0;
	padding: 2px 8px;
	text-align: center;
}

.gowidget tbody th, .gowidget tbody td
{
	color: ${Main.Text};
	background: ${Main.Background};
	border: 1px solid ${Main.Border};
}

.gowidget thead th, .gowidget thead td
{
	color: ${Good.Text};
	background: ${Good.Background};
	border: 1px solid ${Good.Border};
	font-size: 12pt;
}

.gowidget tfoot th, .gowidget tfoot td
{
	color: ${Good.Text};
	border-top: 1px solid ${Good.Border};
}

.gowidget a:link, .gowidget a:hover, .gowidget a:active, .gowidget a:visited
{
	color: ${Main.Text};
	text-decoration: none;
}

.gowidget a:hover
{
	text-decoration: underline;
}

.gowidget thead a:link, .gowidget thead a:hover, .gowidget thead a:active, .gowidget thead a:visited
{
	font-weight: bold;
	color: ${Good.Text};
}

.gowidget th
{
	font-weight: bold;
}

.gowidget tfoot td
{
	font-size: 8pt;
}

.gowidget .tiny
{
	width: 100px;
	display: inline-block;
	font-size: 8pt;
	text-align: left;
}

.gowidget .tiny:first-child
{
	text-align: right;
}


.gowidget .tiny a
{
	padding: 0 4px;
}

.gowidget tbody tr th:first-child
{
	text-align: right;
}
</style>
<script type="text/javascript">
function expand(element, index) {
	table = element.parentNode.parentNode.parentNode.parentNode.parentNode;
	table.tBodies[index].style.display = "table-row-group";
	table.tBodies[1-index].style.display = "none";
}
</script>
`

type Pallete struct {
	Text, Text2, Border, Light, Background string
}

type ColorScheme struct {
	Main, Good, Warn, Bad Pallete
}

var widgetColors = ColorScheme{
	Main: Pallete{"#2E733E", "#30663C", "#1E602D", "#8BDD9D", "#B3DDBC"},
	Good: Pallete{"#26585C", "#274E52", "#19494E", "#89D1D8", "#AFD4D8"},
	Warn: Pallete{"#98683D", "#866140", "#805128", "#E6B991", "#E6CFBA"},
	Bad:  Pallete{"#98463D", "#864640", "#803028", "#E69991", "#E6BEBA"},
}

var widgetTemplate = template.MustParse(``+
`<table class="gowidget">
	<thead>
		<tr>
			<th colspan="3">
				<a href="{HomeURL}">{Name}</a> - 5/5
			</th>
		</tr>
	</thead>
	<tfoot>
		<tr>
			<td colspan=3>
				<div class="tiny">
					<a href="#" onclick="expand(this, 0);return false">Project Info</a>
				</div>
				<div class="tiny">
					<a href="#" onclick="expand(this, 1);return false">Build Status</a>
				</div>
			</td>
		</tr>
	</tfoot>
	<tbody>
		<tr>
			<td>
				Rating: {Rating} (<a href="/hook/plusone/{ID}">+</a>)
			</td>
			<td>
				<a href="/hook/wontbuild/{ID}">Won&#39;t Build</a></span>
			</td>
		</tr>
		<tr>
			<td>
				<a href="{SourceURL}">Source Code</a></span>
			</td>
			<td>
				<a href="{BugURL}">Report Bug</a></span>
			</td>
		</tr>
	</tbody>
	<tbody style="display: none">
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
`, nil)

func (w *Widget) Execute(out io.Writer) os.Error {
	if len(widgetStatic) == 0 {
		buf := bytes.NewBuffer(nil)
		static := template.New(nil)
		static.SetDelims("${", "}")
		err := static.Parse(widgetStaticTemplate)
		if err == nil {
			err = static.Execute(buf, widgetColors)
		}
		if err != nil {
			fmt.Fprintf(buf, "<b>Error</b>: %s<br/>", err)
		}
		widgetStatic = buf.String()
	}
	fmt.Fprint(out, widgetStatic)
	return widgetTemplate.Execute(out, w)
}

// ExecuteString returns a string safe to embed in a single-quoted string.  If
// the underlying call to Execute has a single quote in it, an error string
// suitable for printing is returned.
func (w *Widget) ExecuteString() string {
	buf := bytes.NewBuffer(nil)
	err := w.Execute(buf)
	if err != nil {
		return "<b>Error:</b> " + html.EscapeString(err.String())
	}
	if bytes.IndexByte(buf.Bytes(), '\'') >= 0 {
		return "<p><b>Error</b>: Banned character found in template</p>"
	}
	return buf.String()
}
