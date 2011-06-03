package widget

import (
	"crypto/md5"
	"fmt"
	"os"

	"appengine"
	"appengine/user"
	"appengine/datastore"
)

type Widget struct {
	ctx            appengine.Context
	key            *datastore.Key
	Name           string
	ID             string
	Owner          string
	CompileTotal   int
	CompileWeek    int
	CompileCheckin int
	CheckinTotal   int
	CheckinWeek    int
	LastCompile    datastore.Time
	LastCheckin    datastore.Time
	StatsUpdated   datastore.Time
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
		ctx: ctx,
		key: datastore.NewKey("Widget", hash, 0, nil),
		Name: name,
		ID: hash,
		Owner: u.Email,
	}
}

func LoadWidget(ctx appengine.Context, id string) (widget *Widget, err os.Error) {
	key := datastore.NewKey("Widget", id, 0, nil)

	widget = new(Widget)
	err = datastore.Get(ctx, key, widget)
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
