package times

import (
	"errors"
	"os"
	"runtime"
	"strings"
)

func (t *Times) GetTimezone() (string, error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return "", errors.New("incorrect host, must be linux or darwin")
	}
	data, err := os.ReadFile("/etc/timezone")
	if err != nil {
		return "", err
	}
	timezonePart := strings.TrimSpace(string(data))
	return timezonePart, nil
}
