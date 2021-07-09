package main

import (
	"fmt"
	"os"

	"github.com/PatrickRudolph/filebin/internal/utils"
)

func main() {

	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <URL> <Username> <Password>\n", os.Args[0])
		return
	}
	u := utils.HTTPFileUploader{
		Url:      os.Args[1],
		Username: os.Args[2],
		Password: os.Args[3],
	}

	fds, err := u.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		for i := range fds {
			fmt.Printf("%d: %v\n", i, fds[i])
		}
	}
}
