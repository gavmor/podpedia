package main

import (
	"fmt"
	"strings"
)

func validateDownloadRequest(url, dest string) error {
	switch {
	case url == "":
		return fmt.Errorf("url required")
	case dest == "":
		return fmt.Errorf("dest required")
	case !strings.HasPrefix(url, "http"):
		return fmt.Errorf("url must be http(s)")
	}
	return nil
}
