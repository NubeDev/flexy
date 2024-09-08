package systeminfo

import (
	"fmt"
	"testing"
)

func TestNewSystem(t *testing.T) {
	s := NewSystem()
	fmt.Println(s.GetHostUniqueID())
	fmt.Println(getPublicIP())
}
