package rqlclient

import "time"

func (inst *Client) GitDownloadAsset(owner, repo, tag, arch, token string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"owner": owner, "repo": repo, "tag": tag, "arch": arch, "token": token}
	return inst.biosCommandRequest(body, "post", "git", "manager.asset", timeout)
}

func (inst *Client) GitListAsset(owner, repo, tag, arch, token string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"owner": owner, "repo": repo, "tag": tag, "arch": arch, "token": token}
	return inst.biosCommandRequest(body, "get", "git", "manager.asset", timeout)
}
