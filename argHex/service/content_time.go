package service

import "time"

// contentStampLayout is fixed-width (nanoseconds always 9 digits) so that a
// plain string sort is a chronological sort even within the same second;
// RFC3339Nano trims trailing zeros and breaks that.
const contentStampLayout = "2006-01-02T15:04:05.000000000Z07:00"

// nowStamp is the single source of the content timestamp format: fixed-width
// RFC3339 in UTC, which is how revisions and the activity log are ordered in
// mongo.
func nowStamp() string {
	return time.Now().UTC().Format(contentStampLayout)
}
