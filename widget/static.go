package widget

import (
	"template"
	"io"
	"bytes"
	"fmt"

	"appengine"
	"appengine/user"
)

type Pallete struct {
	Text, Text2, Border, Light, Background string
}

type ColorScheme struct {
	Main, Good, Warn, Bad Pallete
}

var gowidgetColors = ColorScheme{
	Main: Pallete{"#2E733E", "#30663C", "#1E602D", "#8BDD9D", "#B3DDBC"},
	Good: Pallete{"#26585C", "#274E52", "#19494E", "#89D1D8", "#AFD4D8"},
	Warn: Pallete{"#98683D", "#866140", "#805128", "#E6B991", "#E6CFBA"},
	Bad:  Pallete{"#98463D", "#864640", "#803028", "#E69991", "#E6BEBA"},
}

var widgetStatic string
var widgetStaticTemplate = `` +
`<style type='text/css'>
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

.gowidget tbody th
{
	width: 70px;
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
	table = element.parentNode.parentNode.parentNode.parentNode;
	table.tBodies[index].style.display = "table-row-group";
	table.tBodies[1-index].style.display = "none";
}
</script>
`

func writeWidgetCSS(w io.Writer) {
	if len(widgetStatic) == 0 {
		buf := bytes.NewBuffer(nil)
		static := template.New(nil)
		static.SetDelims("${", "}")
		err := static.Parse(widgetStaticTemplate)
		if err == nil {
			err = static.Execute(buf, gowidgetColors)
		}
		if err != nil {
			fmt.Fprintf(buf, "<b>Error</b>: %s<br/>", err)
		}
		widgetStatic = buf.String()
	}
	fmt.Fprint(w, widgetStatic)
}

var commonStatic string
var commonStaticTemplate = `` +
`<style type='text/css'>
body
{
	margin-top: 30px;
}

#header
{
	position: absolute;
	top: 0px;
	right: 0px;
	left: 0px;
	margin: 0px;
	padding: 0px;
	text-align: right;
	color: ${Main.Text};
	background-color: ${Main.Background};
	border-bottom: 1px solid ${Main.Border};
}

#header #userinfo
{
	padding: 5px 15px;
	font-size: 15px;
}

#header a:link, #header a:hover, #header a:active, #header a:visited
{
	color: ${Main.Text};
	text-decoration: none;
}

#header a:hover
{
	text-decoration: underline;
}

.leaderBoard
{
  counter-reset: widget;
}

.leaderBoard
{
	margin: 5px;
	border-collapse: collapse;
	border-spacing: 0;
}

.leaderBoard th, .leaderboard td
{
	font-weight: normal;
	font-size: 12pt;
	margin: 0;
	padding: 2px 8px;
	text-align: center;
}

.leaderBoard tbody th, .leaderBoard tbody td
{
	color: ${Main.Text};
	background: ${Main.Background};
	border: 1px solid ${Main.Border};
}

.leaderBoard thead th
{
	color: ${Good.Text};
	background: ${Good.Background};
	border: 1px solid ${Good.Border};
	font-size: 14pt;
}

.leaderBoard th
{
	font-weight: bold;
}

.leaderBoard a:link, .leaderBoard a:hover, .leaderBoard a:active, .leaderBoard a:visited
{
	color: ${Main.Text};
	text-decoration: none;
}

.leaderBoard a:hover
{
	text-decoration: underline;
}

.leaderBoard .topWidget:before
{
  counter-increment: widget;
  content: "#" counter(widget);
}

.leaderBoard .right
{
	text-align: right;
}

.leaderboard .left
{
	text-align: left;
}
</style>
`

func commonCSS() string {
	if len(commonStatic) == 0 {
		buf := bytes.NewBuffer(nil)
		static := template.New(nil)
		static.SetDelims("${", "}")
		err := static.Parse(commonStaticTemplate)
		if err == nil {
			err = static.Execute(buf, gowidgetColors)
		}
		if err != nil {
			buf.Truncate(0)
			fmt.Fprintf(buf, "<b>Error</b>: %s<br/>", err)
		}
		commonStatic = buf.String()
	}
	return commonStatic
}

var headerTemplate = ``+
`<div id='header'>
	<div id='userinfo'>
{.section User}
		{Email|html}
		| <a href="/logout">Log Out</a> 
		| <a href="/widget/list">My Projects</a>
{.or}
		<a href="/login">Log In</a>
{.end}
		| <a href="/leaderboard">Popular Projects</a>
	</div>
</div>
`

type headerData struct {
	User *user.User
}

func header(ctx appengine.Context) string {
	page, err := template.Parse(headerTemplate, nil)
	if err != nil {
		return fmt.Sprintf("<b>Error</b>: %s<br/>", err)
	}

	data := &headerData{
		User: user.Current(ctx),
	}

	buf := bytes.NewBuffer(nil)
	err = page.Execute(buf, data)
	if err != nil {
		return fmt.Sprintf("<b>Error</b>: %s<br/>", err)
	}

	return buf.String()
}
