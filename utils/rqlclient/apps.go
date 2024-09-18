package rqlclient

import "time"

// BiosInstallApp installs an app with the given name and version on the specified client
func (inst *Client) BiosInstallApp(appName, version, appID string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": appName, "version": version, "appID": appID}
	return inst.biosCommandRequest(body, "post", "apps", "manager.install", timeout)
}

// BiosUninstallApp uninstalls an app with the given name and version on the specified client
func (inst *Client) BiosUninstallApp(appName, version, appID string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": appName, "version": version, "appID": appID}
	return inst.biosCommandRequest(body, "post", "apps", "manager.uninstall", timeout)
}

type Message struct {
	Message string `json:"message"`
}

// BiosInstalledApps retrieves a list of installed apps on the client
func (inst *Client) BiosInstalledApps(timeout time.Duration) (interface{}, error) {
	body := map[string]string{"body": ""}
	return inst.biosCommandRequest(body, "get", "apps", "manager.installed", timeout)
}

// BiosLibraryApps retrieves a list of available apps in the library on the client
func (inst *Client) BiosLibraryApps(timeout time.Duration) (interface{}, error) {
	body := map[string]string{"body": ""}
	return inst.biosCommandRequest(body, "get", "apps", "manager.library", timeout)
}
