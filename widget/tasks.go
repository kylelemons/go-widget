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
		err os.Error
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
		if e != nil { done <- e; return }

		// Delete old ones
		var i int
		for i = len(comps)-1; i >= 0; i-- {
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
		if e != nil { done <- e; return }

		// Delete old ones
		var i int
		for i = len(commits)-1; i >= 0; i-- {
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
				if !dopen { continue }
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
				if !copen { continue }
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
		wCpy := &Widget{
			ctx: ctx,
			key: w.key,
			Name: w.Name,
			ID:   w.ID,
			Owner: w.Owner,
			CompileTotal: w.CompileTotal,
			CheckinTotal: w.CheckinTotal,
			StatsUpdated: start,
		}
		counts[w.ID] = wCpy
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
