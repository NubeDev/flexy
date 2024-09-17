package githubdownloader

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
)

// RepoAsset holds information about a repository asset.
type RepoAsset struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Tag   string `json:"tag"`
	Arch  string `json:"arch"`
	Token string `json:"token"`
}

// Asset represents a release asset with additional metadata.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ReleaseTag         string `json:"release_tag"`
}

// GitHubDownloader is a client for downloading GitHub releases.
type GitHubDownloader struct {
	client          *github.Client
	gitDownloadPath string
	ctx             context.Context
}

// New creates a new GitHubDownloader instance.
func New(token, gitDownloadPath string) *GitHubDownloader {
	ctx := context.Background()
	var httpClient *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(ctx, ts)
	}
	client := github.NewClient(httpClient)

	return &GitHubDownloader{
		client:          client,
		gitDownloadPath: gitDownloadPath,
		ctx:             ctx,
	}
}

// UpdateToken updates the authentication token and recreates the GitHub client.
func (gd *GitHubDownloader) UpdateToken(token string) {
	var httpClient *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(gd.ctx, ts)
	}
	gd.client = github.NewClient(httpClient)
}

// UpdateDownloadPath updates the download path for assets.
func (gd *GitHubDownloader) UpdateDownloadPath(path string) {
	gd.gitDownloadPath = path
}

// DownloadRelease downloads a GitHub release asset matching the specified architecture.
func (gd *GitHubDownloader) DownloadRelease(owner, repo, tag, arch string) error {
	release, _, err := gd.client.Repositories.GetReleaseByTag(gd.ctx, owner, repo, tag)
	if err != nil {
		return err
	}

	// Get assets for the release.
	opt := &github.ListOptions{PerPage: 100}
	var allAssets []*github.ReleaseAsset

	for {
		assets, resp, err := gd.client.Repositories.ListReleaseAssets(gd.ctx, owner, repo, release.GetID(), opt)
		if err != nil {
			return err
		}
		allAssets = append(allAssets, assets...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Find the asset matching the architecture.
	var assetFound *github.ReleaseAsset
	for _, asset := range allAssets {
		if arch != "" && !strings.Contains(asset.GetName(), arch) {
			continue
		}
		assetFound = asset
		break
	}

	if assetFound == nil {
		return errors.New("asset not found")
	}

	// Create the directory if it doesn't exist.
	if _, err := os.Stat(gd.gitDownloadPath); os.IsNotExist(err) {
		err = os.MkdirAll(gd.gitDownloadPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Download the asset.
	rc, redirectURL, err := gd.client.Repositories.DownloadReleaseAsset(gd.ctx, owner, repo, assetFound.GetID(), http.DefaultClient)
	if err != nil {
		return err
	}

	var reader io.ReadCloser
	if rc != nil {
		reader = rc
	} else if redirectURL != "" {
		// Fallback for assets that require redirection.
		resp, err := http.Get(redirectURL)
		if err != nil {
			return err
		}
		reader = resp.Body
	} else {
		return errors.New("failed to download asset")
	}
	defer reader.Close()

	// Save the file.
	outputPath := filepath.Join(gd.gitDownloadPath, assetFound.GetName())
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	return nil
}

// ListAllAssets lists all assets across all releases.
func (gd *GitHubDownloader) ListAllAssets(owner, repo string) ([]Asset, error) {
	opt := &github.ListOptions{PerPage: 100}
	var allReleases []*github.RepositoryRelease

	for {
		releases, resp, err := gd.client.Repositories.ListReleases(gd.ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allReleases = append(allReleases, releases...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var allAssets []Asset
	for _, release := range allReleases {
		releaseAssets, err := gd.listAllAssetsForRelease(owner, repo, release.GetID())
		if err != nil {
			return nil, err
		}
		for _, asset := range releaseAssets {
			allAssets = append(allAssets, Asset{
				Name:               asset.GetName(),
				BrowserDownloadURL: asset.GetBrowserDownloadURL(),
				ReleaseTag:         release.GetTagName(),
			})
		}
	}

	return allAssets, nil
}

// ListAssetsByVersion lists all assets for a specific release tag.
func (gd *GitHubDownloader) ListAssetsByVersion(owner, repo, tag string) ([]Asset, error) {
	release, _, err := gd.client.Repositories.GetReleaseByTag(gd.ctx, owner, repo, tag)
	if err != nil {
		return nil, err
	}

	releaseAssets, err := gd.listAllAssetsForRelease(owner, repo, release.GetID())
	if err != nil {
		return nil, err
	}

	var assets []Asset
	for _, asset := range releaseAssets {
		assets = append(assets, Asset{
			Name:               asset.GetName(),
			BrowserDownloadURL: asset.GetBrowserDownloadURL(),
			ReleaseTag:         release.GetTagName(),
		})
	}

	return assets, nil
}

// ListAssetsByArch lists all assets across all releases that match the given architecture.
func (gd *GitHubDownloader) ListAssetsByArch(owner, repo, arch string) ([]Asset, error) {
	allAssets, err := gd.ListAllAssets(owner, repo)
	if err != nil {
		return nil, err
	}

	var filteredAssets []Asset
	for _, asset := range allAssets {
		if strings.Contains(asset.Name, arch) {
			filteredAssets = append(filteredAssets, asset)
		}
	}

	return filteredAssets, nil
}

// Helper method to list all assets for a release.
func (gd *GitHubDownloader) listAllAssetsForRelease(owner, repo string, releaseID int64) ([]*github.ReleaseAsset, error) {
	opt := &github.ListOptions{PerPage: 100}
	var allAssets []*github.ReleaseAsset

	for {
		assets, resp, err := gd.client.Repositories.ListReleaseAssets(gd.ctx, owner, repo, releaseID, opt)
		if err != nil {
			return nil, err
		}
		allAssets = append(allAssets, assets...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allAssets, nil
}
