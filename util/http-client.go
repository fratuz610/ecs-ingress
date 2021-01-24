package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// HTTPDownloadFile downloads a file via HTTP with an optional header specified
func HTTPDownloadFile(url string, header string) ([]byte, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("Invalid URL: %v", err)
	}

	if header != "" {

		headerList, err := ParseHeader(header)

		if err != nil {
			return nil, fmt.Errorf("Invalid header %v", err)
		}

		req.Header.Add(headerList[0], headerList[1])
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Http Call failed %v", err)
	}

	buf, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("Http Call failed: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Http Call failed with code %v", resp.StatusCode)
	}

	return buf, nil
}

// ParseHeader exported
func ParseHeader(header string) ([]string, error) {

	headerList := strings.Split(header, ":")

	if len(headerList) != 2 {
		return nil, fmt.Errorf("Invalid header %v", header)
	}

	headerName := strings.Trim(headerList[0], " ")
	headerValue := strings.Trim(headerList[1], " ")

	return []string{headerName, headerValue}, nil
}
