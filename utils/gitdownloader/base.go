package githubdownloader

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
	ZipDownloadURL     string `json:"zip_download_url"`
	AssetID            int64  `json:"asset_id"`
	ReleaseTag         string `json:"release_tag"`
	Version            string `json:"version"`
	Arch               string `json:"arch"`
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

// DownloadRelease downloads the specified release zip (using the zipball URL),
// unzips it, and rezips it without the outer folder. It saves the final zip file
// to the provided destination directory, using the release name from GitHub.
func (gd *GitHubDownloader) DownloadRelease(url, destinationDir, releaseName string) error {
	// Ensure the destination directory is not empty
	if destinationDir == "" {
		return fmt.Errorf("destination directory cannot be empty")
	}

	// Create a temporary zip file for the download
	tempZipFile, err := os.CreateTemp("", "github_release_*.zip")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tempZipFile.Name()) // Ensure the temp file is removed afterwards
	defer tempZipFile.Close()

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Use the GitHub client to execute the request
	resp, err := gd.client.Client().Do(req)
	if err != nil {
		return fmt.Errorf("error making the request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	// Write the response body (the zip file) to the temporary file
	_, err = io.Copy(tempZipFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to temp file: %w", err)
	}

	// Rewind the temp file for reading
	tempZipFile.Seek(0, 0)

	// Unzip the contents to a temporary directory, stripping the outer folder
	tempDir, err := os.MkdirTemp("", "github_release_unzip_*")
	if err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	err = unzipWithoutOuterFolder(tempZipFile.Name(), tempDir)
	if err != nil {
		return fmt.Errorf("error unzipping file: %w", err)
	}

	// Create the final zip file name from the release name
	finalZipName := fmt.Sprintf("%s.zip", releaseName)
	finalZipPath := filepath.Join(destinationDir, finalZipName)

	// Rezip the contents directly to the destination
	err = zipDirectory(tempDir, finalZipPath)
	if err != nil {
		return fmt.Errorf("error creating final zip file: %w", err)
	}

	fmt.Printf("Successfully downloaded and re-zipped the release to %s\n", finalZipPath)
	return nil
}

// unzipWithoutOuterFolder extracts a zip file to the given destination directory,
// stripping the first path component (the outer folder).
func unzipWithoutOuterFolder(zipFile, dest string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Skip any leading slashes
		fpath := strings.TrimLeft(f.Name, "/")

		// Split the path to remove the outer folder
		parts := strings.SplitN(fpath, "/", 2)
		if len(parts) == 2 {
			fpath = parts[1]
		} else {
			// Skip entries that don't have an inner path
			continue
		}

		fpath = filepath.Join(dest, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// zipDirectory compresses the contents of a directory into a zip file.
func zipDirectory(srcDir, zipFile string) error {
	zipOut, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("error creating zip file: %w", err)
	}
	defer zipOut.Close()

	zipWriter := zip.NewWriter(zipOut)
	defer zipWriter.Close()

	// Walk through the directory and add files to the zip
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the zip header
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip adding directories to the zip
			return nil
		}

		zipFileHeader, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		zipFileHeader.Name = relPath
		zipFileHeader.Method = zip.Deflate

		// Write the file to the zip
		writer, err := zipWriter.CreateHeader(zipFileHeader)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error while zipping directory: %w", err)
	}

	return nil
}

func (gd *GitHubDownloader) DownloadReleaseByArchVersion(owner, repo, version, arch, destinationDir string, opts *github.ListOptions) error {
	assets, err := gd.ListAllAssets(owner, repo, opts)
	if err != nil {
		return err
	}
	var archMatch bool
	var versionMatch bool
	for _, asset := range assets {
		if asset.Arch == arch {
			archMatch = true
			if asset.Version == version {
				versionMatch = true
				err := gd.DownloadRelease(asset.ZipDownloadURL, destinationDir, asset.Name)
				if err != nil {
					return err
				}
				return nil
			}

		}
	}
	if !archMatch {
		return fmt.Errorf("%s is not a valid version", arch)
	}
	if !versionMatch {
		return fmt.Errorf("%s is not a valid version", version)
	}
	return nil

}

// ListAllAssets lists all assets across all releases.
func (gd *GitHubDownloader) ListAllAssets(owner, repo string, opts *github.ListOptions) ([]Asset, error) {
	if opts == nil {
		opts = &github.ListOptions{PerPage: 100}
	}
	releases, _, err := gd.client.Repositories.ListReleases(gd.ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	var allAssets []Asset
	versionRegex := regexp.MustCompile(`-v(\d+\.\d+\.\d+)`)
	for _, asset := range releases {
		assetName := asset.GetName()
		// Extract version and architecture from the asset name
		version := ""
		arch := ""
		// Extract version
		versionMatch := versionRegex.FindStringSubmatch(assetName)
		if len(versionMatch) > 1 {
			version = versionMatch[1] // Extract the version number
		}

		// Extract architecture (only amd64 and armv7 for now)
		if strings.Contains(assetName, "amd64") {
			arch = "amd64"
		} else if strings.Contains(assetName, "armv7") {
			arch = "armv7"
		}

		allAssets = append(allAssets, Asset{
			Name:               asset.GetName(),
			BrowserDownloadURL: asset.GetAssetsURL(),
			ZipDownloadURL:     asset.GetZipballURL(),
			AssetID:            asset.GetID(),
			ReleaseTag:         asset.GetTagName(),
			Version:            fmt.Sprintf("v%s", version),
			Arch:               arch,
		})
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
	allAssets, err := gd.ListAllAssets(owner, repo, nil)
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
