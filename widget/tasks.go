package widget

import (
	"fmt"
	"http"
	"os"
	"sync"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
)

func taskSummary(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	start := now()
	lastweek := now() - datastore.SecondsToTime(7 * 24 * 60 * 60)

	ctx := appengine.NewContext(r)

	var widgets []*Widget
	widgets, err = LoadAllWidgets(ctx)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	var nwidget, ncompile, ncheckin, narch, nerr int

	update := func(widget *Widget) {
		defer wg.Done()
		var err os.Error

		widget.CompileWeek = 0
		widget.CompileCheckin = 0
		widget.CheckinWeek = 0

		var comms []*Commit
		comms, err = LoadCommits(ctx, widget.ID)
		for i, comm := range comms {
			if !comm.Counted {
				widget.CheckinTotal++
				comm.Counted = true
				err = comm.Commit()
				if err != nil {
					ctx.Logf("Error: %s\n", err)
					nerr++
					continue
				}
			}
			if comm.Time > lastweek {
				widget.CheckinWeek++
			} // TODO(kevlar): Delete older than a week
			if i == 0 {
				widget.LastCheckin = comm.Time
			}
			ncheckin++
		}

		var comps []*Compilation
		comps, err = LoadCompilations(ctx, widget.ID)
		for i, comp := range comps {
			if !comp.Counted {
				widget.CompileTotal++
				comp.Counted = true
				err = comp.Commit()
				if err != nil {
					ctx.Logf("Error: %s\n", err)
					nerr++
					continue
				}
			}
			if comp.Time > lastweek {
				widget.CompileWeek++
			} // TODO(kevlar): Delete older than a week
			if comp.Time > widget.LastCheckin {
				widget.CompileCheckin++
			}
			if i == 0 {
				widget.LastCompile = comp.Time
			}
			ncompile++
		}

		var archs []*Architecture
		archs, err = LoadArchitectures(ctx, widget.ID)
		for _, arch := range archs {
			if arch.Time < widget.LastCheckin {
				err = arch.Delete()
				if err != nil {
					ctx.Logf("Error: %s\n", err)
					nerr++
					continue
				}
			}
			narch++
		}

		widget.StatsUpdated = start
		err = widget.Commit()
		if err != nil {
			ctx.Logf("Error: %s\n", err)
			nerr++
			return
		}
		nwidget++
	}

	for _, widget := range widgets {
		wg.Add(1)
		go update(widget)
	}
	wg.Wait()

	ctx.Logf("Processed %d widgets", nwidget)
	ctx.Logf("Processed %d checkins", ncheckin)
	ctx.Logf("Processed %d compiles", ncompile)
	ctx.Logf("Processed %d architectures", narch)
	ctx.Logf("Encountered %d errors", nerr)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK")
}

func queueTask(w http.ResponseWriter, r *http.Request) {
	var err os.Error
	ctx := appengine.NewContext(r)

	task := taskqueue.NewPOSTTask("/task/test", map[string][]string{
		"key": {"list", "of", "values"},
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
