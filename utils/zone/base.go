package zone

import (
	"github.com/thlib/go-timezone-local/tzlocal"
	"strings"
	"zgo.at/tz"
)

type Zone interface {
	Timezone() string
	Timezones() []string
	TimezonesByCountry() map[string][]string
}

type zone struct{}

func New() Zone {
	return &zone{}
}

func (z zone) TimezonesByCountry() map[string][]string {
	return timezonesByCountry()
}

func (z zone) Timezones() []string {
	var timezones []string
	for _, zone := range tz.Zones {
		timezones = append(timezones, zone.Zone)
	}
	return timezones
}

func timezonesByCountry() map[string][]string {
	timezones := make(map[string][]string)
	for _, zone := range tz.Zones {
		parts := strings.Split(zone.Zone, "/")
		if len(parts) < 2 {
			continue
		}
		country := parts[0]
		timezones[country] = append(timezones[country], zone.Zone)
	}
	return timezones
}

func (z zone) Timezone() string {
	tzname, _ := tzlocal.RuntimeTZ()
	return tzname

}
