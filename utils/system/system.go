package systeminfo

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/zone"
	"github.com/jackpal/gateway"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"math"
	"net"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// System provides methods to retrieve various system-level information
type System interface {
	// GetIP returns the IP address of the system.
	GetIP() string

	// GetUptime returns the system uptime.
	GetUptime() string

	// GetGateway returns the gateway of the system.
	GetGateway() string

	// GetInternetIP retrieves the public IP address of the system.
	GetInternetIP() (*PublicIP, error)

	// GetSystemTime returns the current system time.
	GetSystemTime() *SystemTime

	// TimezoneInfo returns information about the system's timezone.
	TimezoneInfo() zone.Zone

	// GetCurrentCPUUsage returns the current CPU usage of the system.
	GetCurrentCPUUsage() string

	// GetCurrentMemoryUsage returns the current memory usage of the system.
	GetCurrentMemoryUsage() string

	// GetMemoryFree returns the amount of free memory in the system.
	GetMemoryFree() string

	// GetTopProcessesByCPUUsage retrieves the top processes by CPU usage.
	GetTopProcessesByCPUUsage(count int) ([]*topProcess, error)

	// GetTopProcessesByMemory retrieves the top processes by memory usage.
	GetTopProcessesByMemory(count int) ([]*topProcess, error)

	// GetHostUniqueID tries to retrieve the unique identifier of the host, such as MAC address or system UUID.
	GetHostUniqueID() (string, error)

	// Info returns general system information. eg BootTime(), GlobalID(), GlobalUUID()
	Info() *Info
}

type Info struct {
	GlobalID   string      `json:"globalID"`
	GlobalUUID string      `json:"globalUUID"`
	IP         string      `json:"ip"`
	Uptime     string      `json:"uptime"`
	BootTime   string      `json:"bootTime"`
	Version    string      `json:"version"`
	SystemTime *SystemTime `json:"systemTime"`
}

func (i *Info) GetGlobalID() string {
	return i.GlobalID
}

func (i *Info) GetGlobalUUID() string {
	return i.GlobalUUID
}

func (i *Info) GetIP() string {
	return i.IP
}

func (i *Info) GetBootTime() string {
	return i.BootTime
}

func (i *Info) GetUptime() string {
	return i.Uptime
}

func (i *Info) GetVersion() string {
	return i.Version
}

func (i *Info) GetSystemTime() *SystemTime {
	if i.SystemTime != nil {
		return nil
	}
	return i.SystemTime
}

type unixSystem struct {
	bootTime time.Time
}

func (s *unixSystem) Info() *Info {
	u, _ := s.GetHostUniqueID()
	return &Info{
		GlobalUUID: u,
		IP:         s.GetIP(),
		SystemTime: s.GetSystemTime(),
		Uptime:     s.GetUptime(),
		BootTime:   s.bootTime.Format("2006-01-02 15:04:05"),
	}
}

func NewSystem() System {
	return &unixSystem{
		bootTime: time.Now(),
	}
}

func (s *unixSystem) GetIP() string {
	// Simplified implementation for the first non-loopback IPv4 address
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return "Error getting IP"
	}
	for _, addr := range adders {
		if inet, ok := addr.(*net.IPNet); ok && !inet.IP.IsLoopback() {
			if inet.IP.To4() != nil {
				return inet.IP.String()
			}
		}
	}
	return "no IP found"
}

func (s *unixSystem) GetUptime() string {
	t, err := uptime()
	if err != nil {
		return fmt.Sprintf("error getting uptime: %v", err)
	}
	return t
}

func (s *unixSystem) TimezoneInfo() zone.Zone {
	return zone.New()
}

func (s *unixSystem) GetGateway() string {
	g, err := gateway.DiscoverGateway()
	if err != nil {
		return ""
	}
	return g.String()
}

func (s *unixSystem) GetInternetIP() (*PublicIP, error) {
	return getPublicIP()
}

type SystemTime struct {
	LocalTime string `json:"localTime"`
	UTCTime   string `json:"utcTime"`
	Timezone  string `json:"timezone"`
}

func (s *SystemTime) GetLocalTime() string {
	return s.LocalTime
}

func (s *SystemTime) GetUTCTime() string {
	return s.UTCTime
}

func (s *SystemTime) GetTimezone() string {
	return s.Timezone
}

func (s *unixSystem) GetSystemTime() *SystemTime {
	localTime := time.Now().Format(time.RFC3339)
	utcTime := time.Now().UTC().Format(time.RFC3339)

	return &SystemTime{
		LocalTime: localTime,
		UTCTime:   utcTime,
		Timezone:  s.TimezoneInfo().Timezone(),
	}
}

func (s *unixSystem) GetCurrentCPUUsage() string {
	// CPU usage for a short interval
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Sprintf("Error getting CPU usage: %v", err)
	}
	if len(percent) > 0 {
		return fmt.Sprintf("%.2f%%", percent[0])
	}
	return "CPU usage not available"
}

func (s *unixSystem) GetCurrentMemoryUsage() string {
	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Sprintf("Error getting memory usage: %v", err)
	}
	return fmt.Sprintf("Used: %v MB, Total: %v MB", v.Used/1024/1024, v.Total/1024/1024)
}

func (s *unixSystem) GetMemoryFree() string {
	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Sprintf("Error getting free memory: %v", err)
	}
	return fmt.Sprintf("%v MB", v.Free/1024/1024)
}

type topProcess struct {
	PID           int32
	Name          string
	CPUPercentage float64 // CPU usage percentage
	CPUString     string  // Formatted CPU usage as a string
	MemoryMB      uint64  // Memory usage in MB
	Memory        string  // Memory usage in MB
}

func (s *unixSystem) GetTopProcessesByMemory(count int) ([]*topProcess, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var topProcesses []*topProcess
	for _, p := range processes {
		cpuPercent, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		name, _ := p.Name()

		topProcesses = append(topProcesses, &topProcess{
			PID:           p.Pid,
			Name:          name,
			CPUPercentage: cpuPercent,
			CPUString:     fmt.Sprintf("%.2f%%", cpuPercent),
			Memory:        prettyByteSize(int(memInfo.RSS)),
			MemoryMB:      memInfo.RSS / 1024 / 1024, // Convert from bytes to MB
		})

	}

	// Sort by memory usage
	sort.Slice(topProcesses, func(i, j int) bool {
		return topProcesses[i].MemoryMB > topProcesses[j].MemoryMB
	})

	// Take top 'count' processes
	if len(topProcesses) > count {
		topProcesses = topProcesses[:count]
	}

	return topProcesses, nil
}

func (s *unixSystem) GetTopProcessesByCPUUsage(count int) ([]*topProcess, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var topProcesses []*topProcess
	for _, p := range processes {
		cpuPercent, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		name, _ := p.Name()

		topProcesses = append(topProcesses, &topProcess{
			PID:           p.Pid,
			Name:          name,
			CPUPercentage: cpuPercent,
			CPUString:     fmt.Sprintf("%.2f%%", cpuPercent),
			Memory:        prettyByteSize(int(memInfo.RSS)),
			MemoryMB:      memInfo.RSS / 1024 / 1024, // Convert from bytes to MB
		})
	}

	// Sort by CPU usage
	sort.Slice(topProcesses, func(i, j int) bool {
		return topProcesses[i].CPUPercentage > topProcesses[j].CPUPercentage
	})

	// Take top 'count' processes
	if len(topProcesses) > count {
		topProcesses = topProcesses[:count]
	}

	return topProcesses, nil
}
func uptime() (string, error) {
	out, err := exec.Command("uptime", "-p").Output()
	if err != nil {
		return "", err
	}
	return removeNewLines(string(out)), nil
}

func removeNewLines(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}

func mbToGB(mb int) float64 {
	return float64(mb) / 1024
}

func prettyByteSize(b int) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func getPublicIP() (*PublicIP, error) {
	resp, err := http.Get("https://api.ipify.org/?format=json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ipStruct *PublicIP
	err = json.NewDecoder(resp.Body).Decode(&ipStruct)
	if err != nil {
		return nil, err
	}

	return ipStruct, nil
}

type PublicIP struct {
	Ip string `json:"ip"`
}
