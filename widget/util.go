package widget

import (
	"crypto/md5"
	"fmt"
	"time"

	"appengine/datastore"
)

func now() datastore.Time {
	return datastore.Time(time.Nanoseconds()/1e3)
}

func timestr(t datastore.Time) string {
	return time.SecondsToLocalTime(int64(t)/1e6).String()
}

func Hashf(format string, args ...interface{}) string {
	sum := md5.New()
	fmt.Fprintf(sum, format, args...)
	return fmt.Sprintf("%X", sum.Sum())
}
