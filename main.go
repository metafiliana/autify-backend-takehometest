package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autify-backend-takehometest/utils"
	"golang.org/x/net/html"
)

func main() {
	// URL to scrape
	//urlToScrape := "https://google.com"
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in main:", r)
		}
	}()

	// get flag from executable input
	haveMetadataFlag := false

	// get executable input for website url to scrape
	var urls []string
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}

		if arg == "--metadata" || arg == "-metadata" {
			haveMetadataFlag = true
			continue
		}

		urls = append(urls, arg)
	}

	for _, url := range urls {
		fmt.Printf("fetching %s ...\n", url)
		err := createPage(url, haveMetadataFlag)
		if err != nil {
			fmt.Println("error while downloading web pages:", err)
			break
		}
		fmt.Println("done fetching.")
		fmt.Println()
	}

}

func createPage(targetUrl string, haveMetadataFlag bool) error {
	urlName := targetUrl
	if len(targetUrl) > 30 {
		// set max url length for file & folder naming
		urlName = targetUrl[0:30]
	}
	urlName = strings.Replace(urlName, "/", "-", -1)

	// Make an HTTP request for HTML
	htmlResp, err := http.Get(targetUrl)
	if err != nil {
		return fmt.Errorf("failed to make HTML request: %v", err)
	}
	defer htmlResp.Body.Close()

	// Check if the HTML request was successful (status code 200)
	if htmlResp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not ok: %v", htmlResp.StatusCode)
	}

	lastFetch := time.Now()
	assetDirectoryPath := fmt.Sprintf("%s/%s", utils.Directory_Asset, urlName)
	filePath := fmt.Sprintf("%s.html", urlName)

	// check if file/directory already exist
	isDirectoryExist, _ := utils.IsPathExists(assetDirectoryPath)
	if isDirectoryExist {
		utils.RemovePath(assetDirectoryPath)
	}

	isHtmlFileExist, _ := utils.IsPathExists(filePath)
	if isHtmlFileExist {
		utils.RemovePath(filePath)
	}

	// Create a directory to save image and style files
	err = os.MkdirAll(assetDirectoryPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Read the HTML content
	htmlContent, err := io.ReadAll(htmlResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read HTML content: %v", err)
	}

	//modify content
	content, numLinks, numImages := modifyContent(targetUrl, string(htmlContent), urlName)
	if err == nil {
		htmlContent = []byte(content)
	}

	htmlFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %v", err)
	}
	defer htmlFile.Close()

	// Write the HTML content to the file
	_, err = htmlFile.Write(htmlContent)
	if err != nil {
		return fmt.Errorf("failed to write content to HTML file: %v", err)
	}

	fmt.Println("HTML content saved to", filePath)
	fmt.Println("HTML content assets saved to", assetDirectoryPath)
	if haveMetadataFlag {
		fmt.Println("Number of links", numLinks)
		fmt.Println("Number of images", numImages)
		fmt.Println("Last fetch", lastFetch.Format(time.RFC1123))
	}

	return nil

}

func modifyContent(baseUrl, htmlContent, urlName string) (string, int, int) {
	var modifiedContent strings.Builder
	numLinks := 0
	numImages := 0

	reader := strings.NewReader(htmlContent)
	tokenizer := html.NewTokenizer(reader)
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		switch tokenType {
		case html.ErrorToken:
			return modifiedContent.String(), numLinks, numImages
		case html.SelfClosingTagToken, html.StartTagToken:
			if token.Data == "a" {
				numLinks++
			}

			if token.Data == "link" {
				modifyContentStyleLink(token, baseUrl, urlName)
			} else if token.Data == "image" || token.Data == "svg" || token.Data == "img" {
				numImages++
				modifyContentFile(token, baseUrl, urlName)
			} else if token.Data == "script" {
				modifyContentFile(token, baseUrl, urlName)
			}

			modifiedContent.WriteString(token.String())
		default:
			modifiedContent.WriteString(token.String())
		}
	}
}

func modifyContentStyleLink(token html.Token, baseUrl, urlName string) {
	// iterate through link tag
	for _, attr := range token.Attr {
		if attr.Key == "rel" && (attr.Val == "stylesheet" || attr.Val == "preload") {
			for key, attr := range token.Attr {
				if attr.Key == "href" {
					filepath, err := downloadAssetFile(utils.ConstructFileUrl(baseUrl, attr.Val), urlName)
					if err == nil {
						token.Attr[key].Val = *filepath
					}
					break
				}
			}
			break
		}
	}
}

func modifyContentFile(token html.Token, baseUrl, urlName string) {
	// iterate through image related tag
	keySrc, keySrcData := 0, 0
	isSrc, isSrcData := false, false
	for key, attr := range token.Attr {
		if attr.Key == "data-src" {
			keySrcData = key
			isSrcData = true
		}
		if attr.Key == "src" {
			keySrc = key
			isSrc = true
		}
		if (attr.Key == "href" || attr.Key == "xlink:href") && attr.Val != "" {
			filepath, err := downloadAssetFile(utils.ConstructFileUrl(baseUrl, attr.Val), urlName)
			if err == nil {
				token.Attr[key].Val = *filepath
			}
			return
		}
	}

	if isSrc && isSrcData {
		filepath, err := downloadAssetFile(utils.ConstructFileUrl(baseUrl, token.Attr[keySrcData].Val), urlName)
		if err == nil {
			token.Attr[keySrc].Val = *filepath
		}

	} else if isSrc && !isSrcData {
		filepath, err := downloadAssetFile(utils.ConstructFileUrl(baseUrl, token.Attr[keySrc].Val), urlName)
		if err == nil {
			token.Attr[keySrc].Val = *filepath
		}
	}
}

func downloadAssetFile(fileUrl, urlName string) (*string, error) {
	fileRsp, err := http.Get(fileUrl)
	if err != nil {
		fmt.Println(fmt.Sprintf("%s error while making request: ", utils.ErrorMessage_SkipDownloadAssetFile), err)
		return nil, err
	}
	defer fileRsp.Body.Close()

	if fileRsp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status code: %v", fileRsp.StatusCode)
		fmt.Println(fmt.Sprintf("%s : ", utils.ErrorMessage_SkipDownloadAssetFile), err)
		return nil, err
	}

	downloadedFileName := utils.GetFileNameFromURL(fileUrl)
	content, err := io.ReadAll(fileRsp.Body)
	if err != nil {
		fmt.Println(fmt.Sprintf("%s error reading content: ", utils.ErrorMessage_SkipDownloadAssetFile), err)
		return nil, err
	}

	filePath := filepath.Join(fmt.Sprintf("%s/%s", utils.Directory_Asset, urlName), downloadedFileName)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(fmt.Sprintf("%s error creating file: ", utils.ErrorMessage_SkipDownloadAssetFile), err)
		return nil, err
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		fmt.Println(fmt.Sprintf("%s error writing to file: ", utils.ErrorMessage_SkipDownloadAssetFile), err)
		return nil, err
	}

	return &filePath, nil
}
