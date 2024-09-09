package times

type Format string

const (
	TimeDate     Format = "2006:01:02 15:04:05"
	TimeDateZone Format = "2006:01:02 15:04:05 MST"
	TimeDateDay  Format = "Mon, Jan 02 2006 15:04:05"
)
