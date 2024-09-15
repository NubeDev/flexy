package rqlclient

import "time"

func (inst *Client) BiosSystemdCommand(serviceName, action, property string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": serviceName, "action": action, "property": property}
	return inst.biosCommandRequest(body, "get", "system", "systemctl", timeout)
}

// BiosInstallApp installs an app with the given name and version on the specified client
func (inst *Client) BiosInstallApp(appName, version string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": appName, "version": version}
	return inst.biosCommandRequest(body, "post", "apps", "install", timeout)
}

// BiosUninstallApp uninstalls an app with the given name and version on the specified client
func (inst *Client) BiosUninstallApp(appName, version string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": appName, "version": version}
	return inst.biosCommandRequest(body, "post", "apps", "uninstall", timeout)
}

type Message struct {
	Message string `json:"message"`
}

// BiosInstalledApps retrieves a list of installed apps on the client
func (inst *Client) BiosInstalledApps(timeout time.Duration) (interface{}, error) {
	body := map[string]string{"body": ""}
	return inst.biosCommandRequest(body, "get", "apps", "installed", timeout)
}

// BiosLibraryApps retrieves a list of available apps in the library on the client
func (inst *Client) BiosLibraryApps(timeout time.Duration) (interface{}, error) {
	body := map[string]string{"body": ""}
	return inst.biosCommandRequest(body, "get", "apps", "library", timeout)
}
