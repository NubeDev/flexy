package systeminfo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type NetworkInterface struct {
	InterfaceName  string `json:"interfaceName,omitempty"`
	IpAddress      string `json:"ipAddress,omitempty"`
	NetMask        string `json:"netMask,omitempty"`
	GatewayAddress string `json:"gatewayAddress,omitempty"`
	SubNet         string `json:"subNet,omitempty"`
	IsActive       bool   `json:"isActive"`
	HasIP          bool   `json:"hasIP"`
	MacAddress     string `json:"macAddress,omitempty"`
	Error          string `json:"error,omitempty"`
}

func getNetworkInterface(getGateway bool) ([]NetworkInterface, error) {
	var netInfo []NetworkInterface

	// Get a list of all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		ifaceStatus := NetworkInterface{
			InterfaceName: iface.Name,
			MacAddress:    iface.HardwareAddr.String(),
			IsActive:      iface.Flags&net.FlagUp != 0,
		}

		// Get addresses associated with the interface
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ifaceStatus.IpAddress = ipNet.IP.String()
					ifaceStatus.HasIP = true
					ifaceStatus.NetMask = net.IP(ipNet.Mask).String() // Convert netmask to string
					maskBits, _ := ipNet.Mask.Size()
					ifaceStatus.SubNet = fmt.Sprintf("%d", maskBits)

					if getGateway && ifaceStatus.HasIP {
						address, err := getGatewayAddress(ifaceStatus.InterfaceName)
						if err != nil {
							ifaceStatus.Error = fmt.Sprintf("error on get gateway: %s", err.Error())
						} else {
							ifaceStatus.GatewayAddress = address
						}
					}

					break // Only considering the first IPv4 address
				}
			}
		}

		netInfo = append(netInfo, ifaceStatus)
	}

	return netInfo, nil
}

func getGatewayAddress(interfaceName string) (string, error) {
	// Run the 'ip route' command
	cmd := exec.Command("ip", "route")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// Scan through the command output
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "default") && strings.Contains(line, interfaceName) {
			// Split the line and extract the gateway IP
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "via" && i+1 < len(parts) {
					return parts[i+1], nil
				}
			}
		}
	}

	return "", errors.New("gateway not found")
}
