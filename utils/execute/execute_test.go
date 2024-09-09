package execute

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	got := New().AddTimeout(1)
	// cd /path/to/your/dir && go run main.go build ctl.yaml name=driver-bacnet desc="My new service"
	//c := got.Run("cd", "/home/aidan/code/go/rubix-rx/data/bios", "&&", "./bios", "build", "ctl.yaml", "name=driver-bacnet", "desc=\"My new service\"")
	c := got.Run("sh", "-c", "(cd /home/aidan/code/go/rubix-rx/data/bios  && ./bios build ctl.yaml name=driver-bacnet desc=aaaa)")

	fmt.Println(c.AsError())
	fmt.Println(c.AsString())
}
