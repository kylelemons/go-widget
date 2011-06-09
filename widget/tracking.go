package widget

import (
	"os"
	"strings"

	"appengine"
	"appengine/datastore"
)

type Countable struct {
	ctx    appengine.Context
	key    *datastore.Key
	Widget *datastore.Key
	Hash   string
	Time   datastore.Time
}

func NewCountable(ctx appengine.Context, name, widgetid, hash string) *Countable {
	widgetid = strings.ToUpper(widgetid)
	hash = strings.ToUpper(hash)
	widgetkey := datastore.NewKey("Widget", widgetid, 0, nil)
	return &Countable{
		ctx: ctx,
		key: datastore.NewKey(name, widgetid+hash, 0, widgetkey),
		Widget: widgetkey,
		Hash: hash,
		Time: now(),
	}
}

func (c *Countable) Commit() (err os.Error) {
	c.key, err = datastore.Put(c.ctx, c.key, c)
	return
}

func (c *Countable) Delete() (err os.Error) {
	err = datastore.Delete(c.ctx, c.key)
	return
}

func (c *Countable) Cache() (err os.Error) {
	// TODO
	return
}

func LoadCountable(ctx appengine.Context, name, widgetid string) (cs []*Countable, err os.Error) {
	parent := datastore.NewKey("Widget", strings.ToUpper(widgetid), 0, nil)

	query := datastore.NewQuery(name)
	query.Filter("Widget =", parent)
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &cs)
	for i, c := range cs {
		c.ctx = ctx
		c.key = k[i]
	}

	return
}

func LoadAllCountable(ctx appengine.Context, name string) (cs []*Countable, err os.Error) {
	query := datastore.NewQuery(name)
	query.Order("-Time")

	var k []*datastore.Key
	k, err = query.GetAll(ctx, &cs)
	for i, c := range cs {
		c.ctx = ctx
		c.key = k[i]
	}

	return
}
