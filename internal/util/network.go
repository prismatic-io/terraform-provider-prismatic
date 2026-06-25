package util

import (
	"fmt"
	"net/http"
	"os"
)

func UploadFile(localPath string, uploadUrl string, contentType string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, uploadUrl, file)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = stat.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > 400 {
		return fmt.Errorf("upload attempt returned an error: %d", resp.StatusCode)
	}

	return nil
}
