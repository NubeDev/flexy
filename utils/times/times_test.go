package times

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	tt := New()

	aa := tt
	aa.SetFormat(TimeDateDay)

	fmt.Println(aa.AsTime())
	fmt.Println(aa.AsString())
	fmt.Println(aa.InLocal())
	fmt.Println(aa.InUTC())

}
