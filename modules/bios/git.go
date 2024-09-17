package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	githubdownloader "github.com/NubeDev/flexy/utils/gitdownloader"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"github.com/nats-io/nats.go"
)

func (s *Service) DecodeGitRepoAsset(m *nats.Msg) (*githubdownloader.RepoAsset, error) {
	var cmd githubdownloader.RepoAsset
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

func (s *Service) gitDownloader(m *nats.Msg) {
	decoded, err := s.DecodeGitRepoAsset(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			s.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

}

func (s *Service) gitDownloadAsset(m *nats.Msg) {
	decoded, err := s.DecodeGitRepoAsset(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			s.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Owner == "" {
		s.handleError(m.Reply, code.InvalidParams, "owner is required")
		return
	}
	if decoded.Repo == "" {
		s.handleError(m.Reply, code.InvalidParams, "repo is required")
		return
	}
	if decoded.Tag == "" {
		s.handleError(m.Reply, code.InvalidParams, "tag is required")
		return
	}
	if decoded.Arch == "" {
		s.handleError(m.Reply, code.InvalidParams, "arch is required")
		return
	}
	if decoded.Token != "" {
		s.githubDownloader.UpdateToken(decoded.Token)
	}
	err = s.githubDownloader.DownloadRelease(decoded.Owner, decoded.Repo, decoded.Tag, decoded.Arch)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error downloading: %s err: %v", decoded.Repo, err))
	} else {
		s.publish(m.Reply, fmt.Sprintf("downlaoded %s%s", decoded.Repo, s.gitDownloadPath), code.SUCCESS)
	}
}

func (s *Service) gitListAllAssets(m *nats.Msg) {
	decoded, err := s.DecodeGitRepoAsset(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			s.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Owner == "" {
		s.handleError(m.Reply, code.InvalidParams, "owner is required")
		return
	}
	if decoded.Repo == "" {
		s.handleError(m.Reply, code.InvalidParams, "repo is required")
		return
	}
	if decoded.Token != "" {
		s.githubDownloader.UpdateToken(decoded.Token)
	}
	pprint.PrintJSON(decoded)
	resp, err := s.githubDownloader.ListAllAssets(decoded.Owner, decoded.Repo)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error installing app: %v", err))
	} else {
		fmt.Println(1111)
		pprint.PrintJSON(resp)
		content, err := json.Marshal(resp)
		if err != nil {
			return
		}
		s.publish(m.Reply, string(content), code.SUCCESS)
	}
}
