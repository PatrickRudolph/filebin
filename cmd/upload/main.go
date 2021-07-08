package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rafaelmartins/filebin/internal/utils"
)

func main() {

	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s <URL> <Username> <Password> <File to upload>\n", os.Args[0])
		return
	}
	u := utils.HTTPFileUploader{
		Url:      os.Args[1],
		Username: os.Args[2],
		Password: os.Args[3],
	}

	url, err := u.Upload(os.Args[4], filepath.Base(os.Args[4]))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	fmt.Printf("URL: %s\n", url)
}
