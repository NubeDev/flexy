package times

import "time"

func (t *Times) Timestamp(ti ...time.Time) string {
	newTime := time.Now()
	if len(ti) > 0 {
		newTime = ti[0]
	}
	return newTime.Format(string(t.format))
}
