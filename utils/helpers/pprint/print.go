package pprint

import (
	"encoding/json"
	"os"
)

func PrintJSON(x interface{}) {
	ioWriter := os.Stdout
	w := json.NewEncoder(ioWriter)
	w.SetIndent("", "    ")
	w.Encode(x)
}
