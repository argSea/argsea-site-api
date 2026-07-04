package service

import "time"

// nowStamp is the single source of the content timestamp format. RFC3339 in UTC
// sorts correctly as a plain string, which is how revisions and the activity
// log are ordered in mongo.
func nowStamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
