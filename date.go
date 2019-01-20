package groundcontrol

import "time"

const DateFormat = "2006-01-02T15:04:05-0700"

func NowFormatted() string {
	return time.Now().Format(DateFormat)
}
