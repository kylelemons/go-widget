package widget

import (
	"time"

	"appengine/datastore"
)

func now() datastore.Time {
	return datastore.Time(time.Nanoseconds()/1e3)
}

func timestr(t datastore.Time) string {
	return time.SecondsToLocalTime(int64(t)/1e6).String()
}
