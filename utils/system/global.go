package systeminfo

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func (s *unixSystem) GetHostUniqueID() (string, error) {
	var serialNumber string
	var err error
	serialNumber, err = machineID()
	if err == nil {
		return serialNumber, nil
	}
	serialNumber, err = getSystemSerialNumber()
	if err == nil {
		return serialNumber, nil
	}
	serialNumber, err = getSerialNumber()
	if err == nil {
		return serialNumber, nil
	}
	serialNumber, err = macAsUUID()
	if err != nil {
		panic("failed to to be able to get a host-uuid from this arch")
		return "", err
	}
	return serialNumber, nil
}

// getSerialNumber should work on ubuntu
func machineID() (string, error) {
	const uuidPath = "/etc/machine-id"
	uuidBytes, err := os.ReadFile(uuidPath)
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(string(uuidBytes))
	return uuid, nil
}

// getSerialNumber should work on ubuntu
func getSystemSerialNumber() (string, error) {
	const uuidPath = "/sys/class/dmi/id/product_uuid"
	uuidBytes, err := os.ReadFile(uuidPath)
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(string(uuidBytes))
	return uuid, nil
}

// getSerialNumber should work on like a raspberry
func getSerialNumber() (string, error) {
	const cpuInfoPath = "/proc/cpuinfo"
	data, err := os.ReadFile(cpuInfoPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "Serial" {
			return fields[2], nil
		}
	}
	return "", fmt.Errorf("serial number not found in %s", cpuInfoPath)
}
func macAsUUID() (string, error) {
	var serialNumber string
	interfaces, err := getNetworkInterface(false)
	if err == nil {
		for _, n := range interfaces {
			if n.HasIP {
				mac := n.MacAddress
				if len(mac) > 10 {
					serialNumber = strings.Replace(mac, ":", "", -1)
					serialNumber = fmt.Sprintf("mac_%s", serialNumber)
					return serialNumber, nil
				}
			}
		}
	}
	return "", errors.New("error on get mac address for serial number")
}
