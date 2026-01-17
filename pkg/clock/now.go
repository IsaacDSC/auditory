package clock

import "time"

var Now = func() time.Time {
	return time.Now().UTC()
}

func SetNow(now time.Time) {
	Now = func() time.Time {
		return now
	}
}
