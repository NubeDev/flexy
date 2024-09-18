package appmanager

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"testing"
)

func TestNewAppManager(t *testing.T) {
	manager, err := NewAppManager("", "")
	if err != nil {
		return
	}

	apps, err := manager.ListInstalledApps()
	if err != nil {
		fmt.Println(err)
		return
	}
	pprint.PrintJSON(apps)
}
