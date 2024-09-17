package githubdownloader

//import (
//	"fmt"
//	"github.com/go-resty/resty/v2"
//	"io"
//	"os"
//	"path/filepath"
//	"regexp"
//	"strings"
//)
//
//package githubdownloader
//
//import (
//"encoding/json"
//"errors"
//"fmt"
//"io"
//"os"
//"path/filepath"
//"regexp"
//"strings"
//
//"github.com/go-resty/resty/v2"
//)
//
//type RepoAsset struct {
//	Owner string `json:"owner"`
//	Repo  string `json:"repo"`
//	Tag   string `json:"tag"`
//	Arch  string `json:"arch"`
//	Token string `json:"token"`
//}
//
//// Release represents a GitHub release.
//type Release struct {
//	ID   int    `json:"id"`
//	Tag  string `json:"tag_name"`
//	Name string `json:"name"`
//}
//
//// Asset represents a release asset.
//type Asset struct {
//	Name               string `json:"name"`
//	BrowserDownloadURL string `json:"browser_download_url"`
//	ReleaseTag         string `json:"release_tag"`
//}
//
//// GitHubDownloader is a client for downloading GitHub releases.
//type GitHubDownloader struct {
//	client          *resty.Client
//	token           string
//	gitDownloadPath string
//}
//
//// New creates a new GitHubDownloader instance.
//func New(token, gitDownloadPath string) *GitHubDownloader {
//	client := resty.New()
//	client.SetBaseURL("https://api.github.com")
//	client.SetHeader("Accept", "application/vnd.github+json")
//	client.SetHeader("User-Agent", "githubdownloader")
//	if token != "" {
//		client.SetHeader("Authorization", fmt.Sprintf("token %s", token))
//	}
//
//	return &GitHubDownloader{
//		client:          client,
//		token:           token,
//		gitDownloadPath: gitDownloadPath,
//	}
//}
//
//func (gd *GitHubDownloader) UpdateToken(token string) {
//	gd.token = token
//}
//
//func (gd *GitHubDownloader) UpdateDownloadPath(path string) {
//	gd.gitDownloadPath = path
//}
//
//// DownloadRelease downloads a GitHub release asset matching the specified architecture.
//func (gd *GitHubDownloader) DownloadRelease(owner, repo, tag, arch string) error {
//	assets, err := gd.ListAssetsByVersion(owner, repo, tag)
//	if err != nil {
//		return err
//	}
//
//	// Find the asset matching the architecture.
//	var assetFound *Asset
//	for _, asset := range assets {
//		if arch != "" && !strings.Contains(asset.Name, arch) {
//			continue
//		}
//		assetFound = &asset
//		break
//	}
//
//	if assetFound == nil {
//		return errors.New("asset not found")
//	}
//
//	// Create the directory if it doesn't exist.
//	if _, err := os.Stat(gd.gitDownloadPath); os.IsNotExist(err) {
//		err = os.MkdirAll(gd.gitDownloadPath, os.ModePerm)
//		if err != nil {
//			return err
//		}
//	}
//
//	// Download the asset.
//	downloadURL := assetFound.BrowserDownloadURL
//
//	downloadClient := resty.New()
//	downloadClient.SetHeader("User-Agent", "githubdownloader")
//	if gd.token != "" {
//		downloadClient.SetHeader("Authorization", fmt.Sprintf("token %s", gd.token))
//	}
//
//	downloadResp, err := downloadClient.R().
//		SetDoNotParseResponse(true).
//		Get(downloadURL)
//
//	if err != nil {
//		return err
//	}
//	if downloadResp.StatusCode() != 200 {
//		return fmt.Errorf("failed to download asset: %s", downloadResp.Status())
//	}
//	defer downloadResp.RawBody().Close()
//
//	// Save the file.
//	outputPath := filepath.Join(gd.gitDownloadPath, assetFound.Name)
//	outFile, err := os.Create(outputPath)
//	if err != nil {
//		return err
//	}
//	defer outFile.Close()
//
//	_, err = io.Copy(outFile, downloadResp.RawBody())
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// ListAllAssets lists all assets across all releases.
//func (gd *GitHubDownloader) ListAllAssets(owner, repo string) ([]Asset, error) {
//	releases, err := gd.listAllReleases(owner, repo)
//	if err != nil {
//		return nil, err
//	}
//
//	var allAssets []Asset
//	for _, release := range releases {
//		assets, err := gd.listAllAssetsForRelease(owner, repo, release.ID)
//		if err != nil {
//			return nil, err
//		}
//		// Add release tag to assets
//		for i := range assets {
//			assets[i].ReleaseTag = release.Tag
//		}
//		allAssets = append(allAssets, assets...)
//	}
//
//	return allAssets, nil
//}
//
//// ListAssetsByVersion lists all assets for a specific release tag.
//func (gd *GitHubDownloader) ListAssetsByVersion(owner, repo, tag string) ([]Asset, error) {
//	release, err := gd.getReleaseByTag(owner, repo, tag)
//	if err != nil {
//		return nil, err
//	}
//
//	assets, err := gd.listAllAssetsForRelease(owner, repo, release.ID)
//	if err != nil {
//		return nil, err
//	}
//
//	// Add release tag to assets
//	for i := range assets {
//		assets[i].ReleaseTag = release.Tag
//	}
//
//	return assets, nil
//}
//
//// ListAssetsByArch lists all assets across all releases that match the given architecture.
//func (gd *GitHubDownloader) ListAssetsByArch(owner, repo, arch string) ([]Asset, error) {
//	allAssets, err := gd.ListAllAssets(owner, repo)
//	if err != nil {
//		return nil, err
//	}
//
//	var filteredAssets []Asset
//	for _, asset := range allAssets {
//		if strings.Contains(asset.Name, arch) {
//			filteredAssets = append(filteredAssets, asset)
//		}
//	}
//
//	return filteredAssets, nil
//}
//
//// Helper method to get a release by tag.
//func (gd *GitHubDownloader) getReleaseByTag(owner, repo, tag string) (*Release, error) {
//	resp, err := gd.client.R().
//		SetPathParams(map[string]string{
//			"owner": owner,
//			"repo":  repo,
//			"tag":   tag,
//		}).
//		Get("/repos/{owner}/{repo}/releases/tags/{tag}")
//
//	if err != nil {
//		return nil, err
//	}
//	if resp.StatusCode() != 200 {
//		return nil, fmt.Errorf("failed to get release: %s", resp.Status())
//	}
//
//	var release Release
//	err = json.Unmarshal(resp.Body(), &release)
//	if err != nil {
//		return nil, err
//	}
//
//	return &release, nil
//}
//
//// Helper method to list all releases, handling pagination.
//func (gd *GitHubDownloader) listAllReleases(owner, repo string) ([]Release, error) {
//	var allReleases []Release
//	perPage := 30
//	page := 1
//
//	for {
//		resp, err := gd.client.R().
//			SetPathParams(map[string]string{
//				"owner": owner,
//				"repo":  repo,
//			}).
//			SetQueryParams(map[string]string{
//				"per_page": fmt.Sprintf("%d", perPage),
//				"page":     fmt.Sprintf("%d", page),
//			}).
//			Get("/repos/{owner}/{repo}/releases")
//
//		if err != nil {
//			return nil, err
//		}
//		if resp.StatusCode() != 200 {
//			return nil, fmt.Errorf("failed to get releases: %s", resp.Status())
//		}
//
//		var releases []Release
//		err = json.Unmarshal(resp.Body(), &releases)
//		if err != nil {
//			return nil, err
//		}
//
//		if len(releases) == 0 {
//			break
//		}
//
//		allReleases = append(allReleases, releases...)
//
//		if !hasNextPage(resp.Header().Get("Link")) {
//			break
//		}
//		page++
//	}
//
//	return allReleases, nil
//}
//
//// Helper method to list all assets for a release, handling pagination.
//func (gd *GitHubDownloader) listAllAssetsForRelease(owner, repo string, releaseID int) ([]Asset, error) {
//	var allAssets []Asset
//	perPage := 30
//	page := 1
//
//	for {
//		resp, err := gd.client.R().
//			SetPathParams(map[string]string{
//				"owner":     owner,
//				"repo":      repo,
//				"releaseId": fmt.Sprintf("%d", releaseID),
//			}).
//			SetQueryParams(map[string]string{
//				"per_page": fmt.Sprintf("%d", perPage),
//				"page":     fmt.Sprintf("%d", page),
//			}).
//			Get("/repos/{owner}/{repo}/releases/{releaseId}/assets")
//
//		if err != nil {
//			return nil, err
//		}
//		if resp.StatusCode() != 200 {
//			return nil, fmt.Errorf("failed to get assets: %s", resp.Status())
//		}
//
//		var assets []Asset
//		err = json.Unmarshal(resp.Body(), &assets)
//		if err != nil {
//			return nil, err
//		}
//
//		if len(assets) == 0 {
//			break
//		}
//
//		allAssets = append(allAssets, assets...)
//
//		if !hasNextPage(resp.Header().Get("Link")) {
//			break
//		}
//		page++
//	}
//
//	return allAssets, nil
//}
//
//// Helper function to check if there is a next page in pagination.
//func hasNextPage(linkHeader string) bool {
//	if linkHeader == "" {
//		return false
//	}
//
//	links := parseLinkHeader(linkHeader)
//	_, hasNext := links["next"]
//	return hasNext
//}
//
//// Helper function to parse the Link header.
//func parseLinkHeader(header string) map[string]string {
//	links := make(map[string]string)
//	re := regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)
//	matches := re.FindAllStringSubmatch(header, -1)
//	for _, match := range matches {
//		if len(match) == 3 {
//			links[match[2]] = match[1]
//		}
//	}
//	return links
//}
