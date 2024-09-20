package main

import "fmt"

func main() {
	// Version variable
	version := "V1.02"
	// ASCII Art for "BIOS"
	asciiArt := `
 ____   ___   ___  ____  
| __ ) / _ \ / _ \| __ ) 
|  _ \| | | | | | |  _ \ 
| |_) | |_| | |_| | |_) |
|____/ \___/ \___/|____/ 
`

	// Print the ASCII art
	fmt.Println(asciiArt)

	// Print the BIOS version
	fmt.Printf("RUBIX-BIOS-%s\n", version)
}
