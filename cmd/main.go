package main

import (
	"os"

	"github.com/JamesBLewis/triplej-playlist-generator/internal"
)

// allow go file to be run locally
func main() {
	err := internal.RunBot()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
