package times

import (
	"github.com/andanhm/go-prettytime"
	"time"
)

type Times struct {
	T        time.Time
	asString string
	format   Format
}

func New(t ...time.Time) *Times {
	newTime := time.Now()
	if len(t) > 0 {
		newTime = t[0]
	}
	return &Times{T: newTime}
}

func (t *Times) SetAsUTC() {
	t.T = t.T.UTC()
}

func (t *Times) InLocal() time.Time {
	return t.T.Local()
}

func (t *Times) InUTC() time.Time {
	return t.T.UTC()
}

func (t *Times) SetTimezone(tz string) error {
	if tz != "" {
		tz = "Australia/Sydney"
	}
	location, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	t.T = t.T.In(location)
	return nil
}

func (t *Times) SetFormat(f Format) {
	if f == "" {
		t.format = TimeDate
	} else {
		t.format = f
	}
}

func (t *Times) AsTime() time.Time {
	return t.T
}

func (t *Times) AsString() string {
	if t.format == "" {
		return t.T.Format(time.RFC3339)
	}
	return t.T.Format(string(t.format))
}

// TimeSince returns in a human-readable format the elapsed time
// eg 12 hours, 12 days
func (t *Times) TimeSince() string {
	if t != nil {
		return prettytime.Format(t.T)
	}
	return "error on formatting"

}

func (t *Times) Parse(input string) *Times {
	dateFormats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"02/01/2006 15:04:05",
		"02-Jan-2006 15:04:05",
		"02/01/06 15:04:05",
		"Jan 02, 2006 15:04:05",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000000Z",
		"02-Jan-2006 15:04:05.000",
		"02 January 2006 15:04:05",
		"2006-01-02 15:04:05 -0700 MST",
		"02-01-2006 15:04:05 -0700 MST",
		"02-01-06 15:04:05 -0700 MST",
		"02-Jan-2006 15:04:05 -0700 MST",
		"2006-01-02T15:04:05.000-07:00",
		"02-Jan-2006 15:04:05.000-07:00",
		"January 02, 2006 15:04:05",
		"02-01-2006 15:04:05.000 -0700 MST",
		"Monday, 02-Jan-06 15:04:05 MST",
		"02 Jan 2006 15:04:05 -0700",
		"Jan 02 2006, 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"02 Jan 06 15:04 MST",
		"Jan-02-06 15:04 MST",
		"Jan/02/06 15:04 MST",
		"Jan/02/2006 15:04:05 MST",
		"Mon, 02-Jan-2006 15:04:05 MST",
		"Mon, 02-Jan-06 15:04 MST",
		"Mon, 02 Jan 2006 15:04 MST",
		"Mon, Jan 02, 2006 3:04 PM MST",
		"Mon, Jan 02, 06 3:04 PM MST",
	}

	for _, format := range dateFormats {
		parsedTime, err := time.Parse(format, input)
		if err == nil {
			t.T = parsedTime
			return t
		}
	}
	return nil
}

// Formatted f "2006-01-02 15:04:05"
func (t *Times) Formatted(f string) string {
	return t.T.Format(f)
}

func (t *Times) GetBeginOfYear() *Times {
	t.T = getBeginOfYear(t.T)
	return t
}

func getBeginOfYear(input time.Time) time.Time {
	y, _, _ := input.Date()
	return time.Date(y, 1, 1, 0, 0, 0, 0, input.Location())
}

func (t *Times) GetBeginOfMonth() *Times {
	y, m, _ := t.T.Date()
	t.T = time.Date(y, m, 1, 0, 0, 0, 0, t.T.Location())
	return t
}

func (t *Times) GetBeginOfDay() *Times {
	y, m, d := t.T.Date()
	t.T = time.Date(y, m, d, 0, 0, 0, 0, t.T.Location())
	return t
}

func (t *Times) GetBeginOfHour() *Times {
	y, m, d := t.T.Date()
	t.T = time.Date(y, m, d, t.T.Hour(), 0, 0, 0, t.T.Location())
	return t
}

func (t *Times) GetLastDayOfYear() *Times {
	y, _, _ := t.T.Date()
	t.T = time.Date(y+1, 1, 1, 0, 0, 0, 0, t.T.Location()).AddDate(0, 0, -1)
	return t
}

func (t *Times) GetLastDayOfMonth() *Times {
	t.T = t.GetBeginOfMonth().T.AddDate(0, 1, -1)
	return t
}

func (t *Times) GetTomorrow() *Times {
	t.T = t.GetBeginOfDay().T.AddDate(0, 0, 1)
	return t
}

func (t *Times) GetYesterday() *Times {
	t.T = t.GetBeginOfDay().T.AddDate(0, 0, -1)
	return t
}

func (t *Times) AddSeconds(seconds int) *Times {
	t.T = t.T.Add(time.Duration(seconds) * time.Second)
	return t
}

func (t *Times) SubtractSeconds(seconds int) *Times {
	t.T = t.T.Add(-time.Duration(seconds) * time.Second)
	return t
}

func (t *Times) AddMinutes(minutes int) *Times {
	t.T = t.T.Add(time.Duration(minutes) * time.Minute)
	return t
}

func (t *Times) AddHour(hours int) *Times {
	t.T = t.T.Add(time.Duration(hours) * time.Hour)
	return t
}

func (t *Times) SubtractHours(hours int) *Times {
	t.T = t.T.Add(-time.Duration(hours) * time.Hour)
	return t
}

func (t *Times) AddDay(days int) *Times {
	t.T = t.T.AddDate(0, 0, days)
	return t
}

func (t *Times) SubtractDay(days int) *Times {
	t.T = t.T.AddDate(0, 0, -days)
	return t
}

func (t *Times) AddMonth(months int) *Times {
	t.T = t.T.AddDate(0, months, 0)
	return t
}

func (t *Times) AddWeeks(weeks int) *Times {
	t.T = t.T.AddDate(0, 0, 7*weeks)
	return t
}

func (t *Times) SubtractWeeks(weeks int) *Times {
	t.T = t.T.AddDate(0, 0, -7*weeks)
	return t
}

func (t *Times) SubtractMonth(months int) *Times {
	t.T = t.T.AddDate(0, -months, 0)
	return t
}

func (t *Times) AddYear(years int) *Times {
	t.T = t.T.AddDate(years, 0, 0)
	return t
}

func (t *Times) SubtractYear(years int) *Times {
	t.T = t.T.AddDate(-years, 0, 0)
	return t
}

func (t *Times) NextDay() *Times {
	t.T = t.T.AddDate(0, 0, 1)
	return t
}

func (t *Times) ParseExcelNumber(days int) *Times {
	start := time.Date(1899, 12, 30, 0, 0, 0, 0, t.T.Location())
	t.T = start.Add(time.Duration(days*24) * time.Hour)
	return t
}
