package githubdownloader

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"testing"
)

func TestDownloadRelease(t *testing.T) {
	owner := "NubeDev"
	repo := "flexy-app"
	tag := "v1.0.1"
	arch := "armv7" // armv7 or amd64.
	token := ""
	dir := "./download"

	downloader := New(token, dir)

	// List all assets across all releases
	allAssets, err := downloader.ListAllAssets(owner, repo)
	if err != nil {
		fmt.Printf("Error listing all assets: %v\n", err)
		return
	}
	pprint.PrintJSON(allAssets)
	//
	//fmt.Println("All available assets across all releases:")
	//for _, asset := range allAssets {
	//	fmt.Printf("- [%s] %s\n", asset.ReleaseTag, asset.Name)
	//}

	// List assets by version
	//assetsByVersion, err := downloader.ListAssetsByVersion(owner, repo, tag)
	//if err != nil {
	//	fmt.Printf("Error listing assets by version: %v\n", err)
	//	return
	//}
	//
	//fmt.Printf("\nAssets for version %s:\n", tag)
	//for _, asset := range assetsByVersion {
	//	fmt.Printf("- %s\n", asset.Name)
	//}
	//
	//// List assets by architecture
	//assetsByArch, err := downloader.ListAssetsByArch(owner, repo, arch)
	//if err != nil {
	//	fmt.Printf("Error listing assets by architecture: %v\n", err)
	//	return
	//}
	//
	//fmt.Printf("\nAssets matching architecture '%s':\n", arch)
	//for _, asset := range assetsByArch {
	//	fmt.Printf("- [%s] %s\n", asset.ReleaseTag, asset.Name)
	//}

	// Download release
	err = downloader.DownloadRelease(owner, repo, tag, arch)
	if err != nil {
		fmt.Printf("\nError downloading release: %v\n", err)
	} else {
		fmt.Println("\nDownload completed successfully.")
	}
}
