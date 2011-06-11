package widget

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"os"
	"template"

	"appengine"
	"appengine/user"
	"appengine/memcache"
	"appengine/datastore"
)

type Widget struct {
	ctx appengine.Context
	key *datastore.Key

	populated bool
	dirty bool

	rating int
	broken int

	builds int
	buildWeek int
	buildHead int
	buildLast datastore.Time

	commits int
	commitWeek int
	commitLast datastore.Time

	Name  string
	ID    string
	Owner string

	HomeURL   string
	BugURL    string
	SourceURL string

	// For leaderboard
	CachedScore int64
	CachedRating int64
}

func (w *Widget) CompileDate() string {
	if !w.populated { w.populate() }
	if w.buildLast == 0 {
		return "never"
	}
	return timestr(w.buildLast)
}

func (w *Widget) CheckinDate() string {
	if !w.populated { w.populate() }
	if w.commitLast == 0 {
		return "never"
	}
	return timestr(w.commitLast)
}

func (w *Widget) CompileElapsed() string {
	if !w.populated { w.populate() }
	if w.buildLast == 0 {
		return "never"
	}
	elapsedHours := int64(now()-w.buildLast) / 1e6 / 60 / 60
	elapsedHours, elapsedDays := elapsedHours%24, elapsedHours/24
	return fmt.Sprintf("%dd %dh", elapsedDays, elapsedHours)
}

func (w *Widget) CheckinElapsed() string {
	if !w.populated { w.populate() }
	if w.commitLast == 0 {
		return "never"
	}
	elapsedHours := int64(now()-w.commitLast) / 1e6 / 60 / 60
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

	hash := Hashf("Owner=%s|Widget=%s", u.Email, name)

	return &Widget{
		ctx:    ctx,
		key:    datastore.NewKey("Widget", hash, 0, nil),
		Name:   name,
		ID:     hash,
		Owner:  u.Email,
		populated: true,
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

func LoadTopWidgets(ctx appengine.Context) (widgets []*Widget, err os.Error) {
	query := datastore.NewQuery("Widget")
	query.Order("-CachedScore")
	query.Order("-CachedRating")
	query.Limit(50)

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

func (w *Widget) Score() (score int) {
	if !w.populated { w.populate() }
	if w.rating >= 5 {
		score++
	}
	if w.broken <= 1 {
		score++
	}
	if w.builds >= 50 {
		score++
	}
	if w.buildHead >= 5 {
		score++
	}
	if len(w.BugURL) > 15 && len(w.SourceURL) > 15 && len(w.HomeURL) > 15 {
		score++
	}
	return
}

func (w *Widget) populate() {
	defer func() {
		if err := recover(); err != nil {
			w.ctx.Logf("populate(): %s", err)
		}
	}()

	chk := func(err os.Error) bool {
		switch err {
		case nil:
			return true
		case datastore.Done:
			return false
		}
		panic(err)
	}

	if w.populated { return }

	cache := make(map[string]interface{})
	if !w.dirty {
		if _, err := memcache.Gob.Get(w.ctx, "widget:"+w.ID, &cache); err != memcache.ErrCacheMiss {
			chk(err)

			w.ctx.Logf("Cache hit: %s", w.ID)
			w.ctx.Logf(" - %v", cache)

			w.rating = cache["Rating"].(int)
			w.broken = cache["Broken"].(int)

			w.builds = cache["Builds"].(int)
			w.buildWeek = cache["BuildWeek"].(int)
			w.buildHead = cache["BuildHead"].(int)
			w.buildLast = datastore.Time(cache["BuildLast"].(int64))

			w.commits = cache["Commits"].(int)
			w.commitWeek = cache["CommitWeek"].(int)
			w.commitLast = datastore.Time(cache["CommitLast"].(int64))

			w.populated = true
			return
		}
	} else {
		w.ctx.Logf("Cache: Widget %s is dirty", w.ID)
	}

	lastweek := now() - datastore.SecondsToTime(7*24*60*60)

	var query *datastore.Query
	var items []*Countable
	var err os.Error

	// Broken
	query = datastore.NewQuery("Broken")
	query.Filter("Widget =", w.key)
	w.broken, err = query.Count(w.ctx)
	chk(err)
	w.ctx.Logf("Widget %s has %d broken", w.ID, w.broken)

	// Rating
	query = datastore.NewQuery("Rating")
	query.Filter("Widget =", w.key)
	w.rating, err = query.Count(w.ctx)
	chk(err)
	w.ctx.Logf("Widget %s has %d rating", w.ID, w.rating)

	// Get Commits
	query = datastore.NewQuery("Commit")
	query.Filter("Widget =", w.key)
	query.Order("-Time")
	w.commits, err = query.Count(w.ctx)
	chk(err)
	w.ctx.Logf("Widget %s has %d commits", w.ID, w.commits)
	query.Filter("Time >", lastweek)
	w.commitWeek, err = query.Count(w.ctx)
	chk(err)
	w.ctx.Logf("Widget %s has %d commits this week", w.ID, w.commitWeek)

	query = datastore.NewQuery("Commit")
	query.Filter("Widget =", w.key)
	query.Order("-Time")
	query.Limit(1)
	_, err = query.GetAll(w.ctx, &items)
	chk(err)
	if len(items) > 0 {
		w.commitLast = items[0].Time
	}
	w.ctx.Logf("Widget %s was committed %d", w.ID, w.commitLast)

	// Get builds
	query = datastore.NewQuery("Build")
	query.Filter("Widget =", w.key)
	query.Order("-Time")
	w.builds, err = query.Count(w.ctx)
	chk(err)
	w.ctx.Logf("Widget %s has %d builds", w.ID, w.builds)
	query.Filter("Time >", lastweek)
	w.buildWeek, err = query.Count(w.ctx)
	chk(err)
	w.ctx.Logf("Widget %s has %d builds this week", w.ID, w.buildWeek)

	query = datastore.NewQuery("Build")
	query.Filter("Widget =", w.key)
	query.Order("-Time")
	query.Limit(1)
	_, err = query.GetAll(w.ctx, &items)
	chk(err)
	if len(items) > 0 {
		w.buildLast = items[0].Time
	}
	if w.commitLast > 0 {
		query.Filter("Time >", w.commitLast)
		w.buildHead, err = query.Count(w.ctx)
		chk(err)
	} else {
		w.buildHead = 0
		w.ctx.Logf("Widget %s has no commits", w.ID)
	}
	w.ctx.Logf("Widget %s has %d builds at HEAD", w.ID, w.buildHead)

	w.populated = true
	w.ctx.Logf("Widget %s populated", w.ID)

	w.CachedRating = int64(w.rating)
	w.CachedScore = int64(w.Score())

	cache["Broken"] = w.broken
	cache["Rating"] = w.rating
	cache["Builds"] = w.builds
	cache["BuildWeek"] = w.buildWeek
	cache["BuildHead"] = w.buildHead
	cache["BuildLast"] = int64(w.buildLast)
	cache["Commits"] = w.commits
	cache["CommitWeek"] = w.commitWeek
	cache["CommitLast"] = int64(w.commitLast)
	err = memcache.Gob.Set(w.ctx, &memcache.Item{
		Key: "widget:"+w.ID,
		Expiration: 12*60*60, // 12 hours (because we do some date calculation in here)
		Object: cache,
	})
	chk(err)
	w.ctx.Logf("Cached: Widget %s", w.ID)
}

func (w *Widget) Rating() int {
	if !w.populated { w.populate() }
	return w.rating
}

func (w *Widget) Broken() int {
	if !w.populated { w.populate() }
	return w.broken
}

func (w *Widget) CompileTotal() int {
	if !w.populated { w.populate() }
	return w.builds
}

func (w *Widget) CompileWeek() int {
	if !w.populated { w.populate() }
	return w.buildWeek
}

func (w *Widget) CompileCheckin() int {
	if !w.populated { w.populate() }
	return w.buildHead
}

func (w *Widget) CheckinTotal() int {
	if !w.populated { w.populate() }
	return w.commits
}

func (w *Widget) CheckinWeek() int {
	if !w.populated { w.populate() }
	return w.commitWeek
}

var widgetTemplate = template.MustParse(``+
	`<table class="gowidget">
	<thead>
		<tr>
			<th colspan="3">
				<a href="{HomeURL}">{Name}</a> - {Score}/5
			</th>
		</tr>
	</thead>
	<tfoot>
		<tr>
			<td colspan=3>
				<a href="#" onclick="expand(this, 0);return false">Project</a>
				-
				<a href="#" onclick="expand(this, 1);return false">Builds</a>
				-
				Powered by <a href="http://go-widget.appspot.com/">Go-Widget</a>
			</td>
		</tr>
	</tfoot>
	<tbody>
		<tr>
			<td>
				Rating: {Rating} (<a href="/hook/plusone/{ID}">+</a>)
			</td>
			<td>
				<a href="/hook/wontbuild/{ID}">Broken</a> ({Broken})</span>
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
			<th>Weekly</th>
			<!--InstallWeek-->
			<td>{CompileWeek}</td>
			<td>{CheckinWeek}</td>
		</tr>
		<tr>
			<th>Total</th>
			<!--InstallTotal-->
			<td>{CompileTotal}</td>
			<td>{CheckinTotal}</td>
		</tr>
		<tr>
			<th>Last</th>
			<!--InstallElapsed-->
			<td>{CompileElapsed}</td>
			<td>{CheckinElapsed}</td>
		</tr>
	</tbody>
</table>
`, nil)

func (w *Widget) Execute(out io.Writer) os.Error {
	writeWidgetCSS(out)
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
