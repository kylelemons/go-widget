package widget

import (
	"fmt"
	"http"
	"os"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
)

func taskSummary(w http.ResponseWriter, r *http.Request) {
	var (
		err  os.Error
		done = make(chan os.Error)

		ctx = appengine.NewContext(r)

		start    = now()
		lastweek = now() - datastore.SecondsToTime(7*24*60*60)

		widgets []*Widget
		comps   []*Compilation
		commits []*Commit
		archs   []*Architecture
	)

	go func() {
		var e os.Error
		widgets, e = LoadAllWidgets(ctx)
		done <- e
	}()

	go func() {
		var e os.Error
		comps, e = LoadAllCompilations(ctx)
		if e != nil {
			done <- e
			return
		}

		// Delete old ones
		var i int
		for i = len(comps) - 1; i >= 0; i-- {
			if comps[i].Time > lastweek {
				break
			}
			comps[i].Delete()
			comps = comps[:i]
		}

		done <- e
	}()

	go func() {
		var e os.Error
		commits, e = LoadAllCommits(ctx)
		if e != nil {
			done <- e
			return
		}

		// Delete old ones
		var i int
		for i = len(commits) - 1; i >= 0; i-- {
			if commits[i].Time > lastweek {
				break
			}
			commits[i].Delete()
			commits = commits[:i]
		}

		done <- e
	}()

	go func() {
		var e os.Error
		archs, e = LoadAllArchitectures(ctx)
		done <- e
	}()

	for i := 0; i < 4; i++ {
		e := <-done
		if e != nil {
			if err == nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusInternalServerError)
			}
			err = e
			fmt.Fprintf(w, "Error loading datastore: %s\n", err)
			ctx.Logf("Error loading datastore: %s", err)
			return
		}
	}

	ctx.Logf("Processing: %d widgets, %d compiles, %d commits, %d architectures",
		len(widgets), len(comps), len(commits), len(archs))

	var (
		counts = make(map[string]*Widget)
		del    = make(chan Item, 100)
		com    = make(chan Item, 100)

		errors, deleted, updated int
	)

	go func() {
		var item Item
		dopen := true
		copen := true
		for copen || dopen {
			select {
			case item, dopen = <-del:
				if !dopen {
					continue
				}
				deleted++
				e := item.Delete()
				if e != nil {
					if err == nil {
						errors++
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusInternalServerError)
					}
					err = e
					fmt.Fprintf(w, "Error deleting item: %s\n", err)
					ctx.Logf("Error deleting item: %s", err)
				}
			case item, copen = <-com:
				if !copen {
					continue
				}
				updated++
				e := item.Commit()
				if e != nil {
					if err == nil {
						errors++
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusInternalServerError)
					}
					err = e
					fmt.Fprintf(w, "Error committing item: %s\n", err)
					ctx.Logf("Error committing item: %s", err)
				}
			}
		}
		done <- err
	}()

	for _, w := range widgets {
		wCpy := *w
		wCpy.CompileWeek = 0
		wCpy.CompileCheckin = 0
		wCpy.CheckinWeek = 0
		wCpy.StatsUpdated = start
		counts[w.ID] = &wCpy
	}

	for _, c := range commits {
		w, ok := counts[c.Widget.StringID()]
		if !ok {
			del <- c
		}
		if !c.Counted {
			w.CheckinTotal++
			c.Counted = true
			com <- c
			w.WontBuilds = 0
		}
		w.CheckinWeek++
		if c.Time > w.LastCheckin {
			w.LastCheckin = c.Time
		}
	}

	for _, c := range comps {
		w, ok := counts[c.Widget.StringID()]
		if !ok {
			del <- c
		}
		if !c.Counted {
			w.CompileTotal++
			c.Counted = true
			com <- c
		}
		w.CompileWeek++
		if c.Time > w.LastCheckin {
			w.CompileCheckin++
		}
		if c.Time > w.LastCompile {
			w.LastCompile = c.Time
		}
	}

	for _, a := range archs {
		w, ok := counts[a.Widget.StringID()]
		if !ok || a.Time < w.LastCheckin {
			del <- a
		}
	}

	for _, w0 := range widgets {
		w := counts[w0.ID]
		if w0.CompileTotal != w.CompileTotal ||
			w0.CompileWeek != w.CompileWeek ||
			w0.CompileCheckin != w.CompileCheckin ||
			w0.CheckinTotal != w.CheckinTotal ||
			w0.CheckinWeek != w.CheckinWeek ||
			w0.LastCompile != w.LastCompile ||
			w0.LastCheckin != w.LastCheckin {
			com <- w
		}
	}

	close(del)
	close(com)
	if err := <-done; err == nil {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "OK")
	}
	ctx.Logf("Updated: %d entries", updated)
	ctx.Logf("Deleted: %d entries", deleted)
	ctx.Logf("Errors: %d entries", errors)
}

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

	optInt := func(any interface{}) int {
		i, _ := any.(int64)
		return int(i)
	}

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
		w.PlusOnes = optInt(widget["PlusOnes"])
		w.WontBuilds = optInt(widget["WontBuilds"])
		w.CompileTotal = optInt(widget["CompileTotal"])
		w.CompileWeek = optInt(widget["CompileWeek"])
		w.CompileCheckin = optInt(widget["CompileCheckin"])
		w.CheckinTotal = optInt(widget["CheckinTotal"])
		w.CheckinWeek = optInt(widget["CheckinWeek"])
		w.LastCompile, _ = widget["LastCompile"].(datastore.Time)
		w.LastCheckin, _ = widget["LastCheckin"].(datastore.Time)
		w.StatsUpdated, _ = widget["StatsUpdated"].(datastore.Time)

		widgets = append(widgets, w)
		go func() {
			var err os.Error
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

// unused
func queueTask(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	ctx := appengine.NewContext(r)

	task := taskqueue.NewPOSTTask("/task/test", map[string][]string{
		"key":  {"list", "of", "values"},
		"this": {"is", "pretty", "sweet"},
	})
	err = taskqueue.Purge(ctx, "default")
	if err != nil {
		http.Error(w, fmt.Sprintf("Purge: %s", err), http.StatusInternalServerError)
		return
	}
	err = taskqueue.Add(ctx, task, "default")
	if err != nil {
		http.Error(w, fmt.Sprintf("Add: %s", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Task queued successfully.")
}
