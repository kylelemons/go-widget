package widget

import (
	"fmt"
	"http"
	"os"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
)

func taskUpgrade(out http.ResponseWriter, r *http.Request) {
	out.Header().Set("Content-Type", "text/plain")

	var (
		ctx  = appengine.NewContext(r)
		done = make(chan os.Error)

		testing = false

		widgets []*Widget
	)

	// Load widgets
	query := datastore.NewQuery("Widget")
	iter := query.Run(ctx)

	for {
		widget := make(datastore.Map)
		key, err := iter.Next(widget)
		if err == datastore.Done {
			break
		} else if err != nil {
			fmt.Fprintf(out, "Next(widget): %s\n", err)
			continue
		}
		w := new(Widget)
		w.ctx = ctx
		w.key = key
		w.Name, _ = widget["Name"].(string)
		w.ID, _ = widget["ID"].(string)
		w.Owner, _ = widget["Owner"].(string)
		w.HomeURL, _ = widget["HomeURL"].(string)
		w.BugURL, _ = widget["BugURL"].(string)
		w.SourceURL, _ = widget["SourceURL"].(string)
		w.CachedScore, _ = widget["CachedScore"].(int64)
		w.CachedRating, _ = widget["CachedRating"].(int64)

		widgets = append(widgets, w)
		go func() {
			var err os.Error
			w.populate()
			if !testing {
				err = w.Commit()
			}
			fmt.Fprintf(out, "Widget: Upgraded %s\n", w.ID)
			if testing {
				fmt.Fprintf(out, "   IN  %#v\n", widget)
				fmt.Fprintf(out, "   OUT %#v\n", w)
			}
			done <- err
		}()
	}

	for _ = range widgets {
		err := <-done
		if err != nil {
			fmt.Fprintf(out, "Commit(widget): %s\n", err)
		}
	}

	fmt.Fprintf(out, "COMPLETE\n")
}

func taskRefresh(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	var widget *Widget

	ctx := appengine.NewContext(r)

	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/", -1)
	if len(path) < 3 {
		http.Error(w, "/task/refresh/{widget} - missing required path segment", http.StatusBadRequest)
		return
	}

	widgetID := path[2]
	if len(widgetID) != 32 {
		http.Error(w, "Invalid widget id: " + widgetID, http.StatusBadRequest)
		return
	}

	if widget, err = LoadWidget(ctx, widgetID); err != nil {
		http.Error(w, "Unknown widget id: " + widgetID, http.StatusBadRequest)
		return
	}
	widget.dirty = true
	widget.populate()
	err = widget.Commit()
	if err != nil {
		ctx.Debugf("update: commit: %s", err)
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK")
}

func refreshWidget(w http.ResponseWriter, r *http.Request, widgetID string) {
	var err os.Error
	ctx := appengine.NewContext(r)

	task := taskqueue.NewPOSTTask("/task/refresh/"+widgetID, nil)
	_, err = taskqueue.Add(ctx, task, "default")
	if err != nil {
		http.Error(w, fmt.Sprintf("Add: %s", err), http.StatusInternalServerError)
		return
	}
	ctx.Debugf("Refresh: Widget %s refresh queued", widgetID)
}
