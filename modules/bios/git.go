package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	githubdownloader "github.com/NubeDev/flexy/utils/gitdownloader"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"github.com/nats-io/nats.go"
)

func (inst *Service) DecodeGitRepoAsset(m *nats.Msg) (*githubdownloader.RepoAsset, error) {
	var cmd githubdownloader.RepoAsset
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	pprint.PrintJSON(cmd)
	return &cmd, nil
}

func (inst *Service) gitDownloadAsset(m *nats.Msg) {
	decoded, err := inst.DecodeGitRepoAsset(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			inst.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Owner == "" {
		inst.handleError(m.Reply, code.InvalidParams, "owner is required")
		return
	}
	if decoded.Repo == "" {
		inst.handleError(m.Reply, code.InvalidParams, "repo is required")
		return
	}
	if decoded.Tag == "" {
		inst.handleError(m.Reply, code.InvalidParams, "tag is required")
		return
	}
	if decoded.Arch == "" {
		inst.handleError(m.Reply, code.InvalidParams, "arch is required")
		return
	}
	if decoded.Token != "" {
		inst.githubDownloader.UpdateToken(decoded.Token)
	}
	err = inst.githubDownloader.DownloadRelease(decoded.Owner, decoded.Repo, decoded.Tag, decoded.Arch)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error downloading: %s err: %v", decoded.Repo, err))
	} else {
		inst.publish(m.Reply, fmt.Sprintf("downlaoded %s%s", decoded.Repo, inst.gitDownloadPath), code.SUCCESS)
	}
}

func (inst *Service) gitListAllAssets(m *nats.Msg) {
	decoded, err := inst.DecodeGitRepoAsset(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			inst.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Owner == "" {
		inst.handleError(m.Reply, code.InvalidParams, "owner is required")
		return
	}
	if decoded.Repo == "" {
		inst.handleError(m.Reply, code.InvalidParams, "repo is required")
		return
	}
	resp, err := inst.githubDownloader.ListAllAssets(decoded.Owner, decoded.Repo)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error installing app: %v", err))
	} else {
		content, err := json.Marshal(resp)
		if err != nil {
			return
		}
		inst.publish(m.Reply, string(content), code.SUCCESS)
	}
}
