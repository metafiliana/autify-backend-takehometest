package utils

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
)

func GetFileNameFromURL(urlString string) string {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}

	return path.Base(parsedURL.Path)
}

func ConstructFileUrl(baseUrl, fileUrl string) string {
	if !strings.HasPrefix(fileUrl, "http") || !strings.HasPrefix(fileUrl, "https") {
		// If the URL is relative, construct the absolute URL
		fileUrl = baseUrl + "/" + fileUrl
	}

	return fileUrl
}

func IsPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func RemovePath(path string) error {
	return os.RemoveAll(path)
}
