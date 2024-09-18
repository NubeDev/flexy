package githubdownloader

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"testing"
)

func TestDownloadRelease(t *testing.T) {
	owner := "NubeDev"
	repo := "flexy-app"
	version := "v1.0.0"
	arch := "amd64" // armv7 or amd64.
	token := ""
	dir := "./download"

	downloader := New(token, dir)

	// List all assets across all releases
	allAssets, err := downloader.ListAllAssets(owner, repo, nil)
	if err != nil {
		fmt.Printf("Error listing all assets: %v\n", err)
		return
	}
	pprint.PrintJSON(allAssets)

	// Download release
	err = downloader.DownloadReleaseByArchVersion(owner, repo, arch, version, dir, nil)

	if err != nil {
		fmt.Printf("\nError downloading release: %v\n", err)
	} else {
		fmt.Println("\nDownload completed successfully.")
	}
}
