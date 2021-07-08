package main

import (
	"fmt"
	"os"

	"github.com/rafaelmartins/filebin/internal/utils"
)

func main() {

	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s <URL> <Username> <Password> <ID to delete>\n", os.Args[0])
		return
	}
	u := utils.HTTPFileUploader{
		Url:      os.Args[1],
		Username: os.Args[2],
		Password: os.Args[3],
	}

	resp, err := u.Delete(os.Args[4])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	fmt.Println(resp)
}
