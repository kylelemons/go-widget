package widget

import (
	"os"
	"strings"

	"appengine"
	"appengine/datastore"
)

type Item interface {
	Commit() os.Error
	Delete() os.Error
}

type Commit struct {
	ctx     appengine.Context
	key     *datastore.Key
	Counted bool
	Widget  *datastore.Key
	Time    datastore.Time
}

func NewCommit(ctx appengine.Context, widgetid string) *Commit {
	widgetid = strings.ToUpper(widgetid)
	widgetkey := datastore.NewKey("Widget", widgetid, 0, nil)
	return &Commit{
		ctx: ctx,
		key: datastore.NewKey("Commit", "", 0, widgetkey),
		Widget: datastore.NewKey("Widget", widgetid, 0, nil),
		Time: now(),
	}
}

func (c *Commit) Commit() (err os.Error) {
	c.key, err = datastore.Put(c.ctx, c.key, c)
	return
}

func (c *Commit) Delete() (err os.Error) {
	err = datastore.Delete(c.ctx, c.key)
	return
}

func LoadCommits(ctx appengine.Context, widgetid string) (comms []*Commit, err os.Error) {
	parent := datastore.NewKey("Widget", strings.ToUpper(widgetid), 0, nil)

	query := datastore.NewQuery("Commit")
	query.Filter("Widget =", parent)
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &comms)
	for i, c := range comms {
		c.ctx = ctx
		c.key = k[i]
	}

	return
}

func LoadAllCommits(ctx appengine.Context) (comms []*Commit, err os.Error) {
	query := datastore.NewQuery("Commit")
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &comms)
	for i, c := range comms {
		c.ctx = ctx
		c.key = k[i]
	}

	return
}

type Compilation struct {
	ctx appengine.Context
	key *datastore.Key
	Counted bool
	Widget  *datastore.Key
	Addr    string
	Time    datastore.Time
}

func NewCompilation(ctx appengine.Context, widgetid, addr string) *Compilation {
	widgetid = strings.ToUpper(widgetid)
	widgetkey := datastore.NewKey("Widget", widgetid, 0, nil)
	return &Compilation{
		ctx: ctx,
		key: datastore.NewKey("Compilation", widgetid+addr, 0, widgetkey),
		Widget: datastore.NewKey("Widget", widgetid, 0, nil),
		Addr: addr,
		Time: now(),
	}
}

func (c *Compilation) Commit() (err os.Error) {
	c.key, err = datastore.Put(c.ctx, c.key, c)
	return
}

func (c *Compilation) Delete() (err os.Error) {
	err = datastore.Delete(c.ctx, c.key)
	return
}

func LoadCompilations(ctx appengine.Context, widgetid string) (comps []*Compilation, err os.Error) {
	parent := datastore.NewKey("Widget", strings.ToUpper(widgetid), 0, nil)

	query := datastore.NewQuery("Compilation")
	query.Filter("Widget =", parent)
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &comps)
	for i, c := range comps {
		c.ctx = ctx
		c.key = k[i]
	}

	return
}

func LoadAllCompilations(ctx appengine.Context) (comps []*Compilation, err os.Error) {
	query := datastore.NewQuery("Compilation")
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &comps)
	for i, c := range comps {
		c.ctx = ctx
		c.key = k[i]
	}

	return
}

type Architecture struct {
	ctx appengine.Context
	key *datastore.Key
	Widget *datastore.Key
	Hash   string
	Time   datastore.Time
}

func NewArchitecture(ctx appengine.Context, widgetid, archhash string) *Architecture {
	widgetid = strings.ToUpper(widgetid)
	archhash = strings.ToUpper(archhash)
	widgetkey := datastore.NewKey("Widget", widgetid, 0, nil)
	return &Architecture{
		ctx: ctx,
		key: datastore.NewKey("Architecture", widgetid+archhash, 0, widgetkey),
		Widget: widgetkey,
		Hash: archhash,
		Time: now(),
	}
}

func (a *Architecture) Commit() (err os.Error) {
	a.key, err = datastore.Put(a.ctx, a.key, a)
	return
}

func (a *Architecture) Delete() (err os.Error) {
	err = datastore.Delete(a.ctx, a.key)
	return
}

func LoadArchitectures(ctx appengine.Context, widgetid string) (ars []*Architecture, err os.Error) {
	parent := datastore.NewKey("Widget", strings.ToUpper(widgetid), 0, nil)

	query := datastore.NewQuery("Architecture")
	query.Filter("Widget =", parent)
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &ars)
	for i, a := range ars {
		a.ctx = ctx
		a.key = k[i]
	}

	return
}

func LoadAllArchitectures(ctx appengine.Context) (ars []*Architecture, err os.Error) {
	query := datastore.NewQuery("Architecture")
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &ars)
	for i, a := range ars {
		a.ctx = ctx
		a.key = k[i]
	}

	return
}
