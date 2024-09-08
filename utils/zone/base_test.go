package zone

import (
	"fmt"
	"testing"
)

func Test_zone_Timezone(t *testing.T) {
	s := New()

	fmt.Println(s.Timezone())
	fmt.Println(s.Timezones())
	fmt.Println(s.TimezonesByCountry())

}
