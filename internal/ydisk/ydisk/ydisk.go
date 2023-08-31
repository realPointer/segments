package ydisk

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
)

type YandexDisk struct {
	token string
}

func NewYandexDisk(token string) *YandexDisk {
	return &YandexDisk{token: token}
}

func (d *YandexDisk) UploadAndReturnDownloadURL(ctx context.Context, name string, data []string) (string, error) {
	if !d.IsAvailable() {
		return "", errors.New("Yandex Disk is not available")
	}

	// Convert []string to CSV format
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, record := range data {
		if err := w.Write([]string{record}); err != nil {
			return "", err
		}
	}
	w.Flush()

	// Upload CSV file to Yandex Disk
	uploadPath, err := d.uploadCSVFile(ctx, name, buf.Bytes())
	if err != nil {
		return "", err
	}
	log.Default().Println(uploadPath)

	// Get download URL for uploaded file
	downloadURL, err := d.getDownloadURL(ctx, uploadPath)
	if err != nil {
		return "", err
	}

	return downloadURL, nil
}

func (d *YandexDisk) IsAvailable() bool {
	resp, err := http.Head("https://cloud-api.yandex.net/v1/disk/")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusMethodNotAllowed
}

func (d *YandexDisk) uploadCSVFile(ctx context.Context, name string, data []byte) (string, error) {
	const baseURL = "https://cloud-api.yandex.net/v1/disk/resources/upload"

	params := url.Values{}
	params.Set("path", name)
	params.Set("overwrite", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "OAuth "+d.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get upload URL")
	}

	var uploadURL struct {
		Href string `json:"href"`
	}

	err = json.NewDecoder(resp.Body).Decode(&uploadURL)
	if err != nil {
		return "", err
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPut, uploadURL.Href, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("failed to upload file")
	}

	return name, nil
}

func (d *YandexDisk) getDownloadURL(ctx context.Context, path string) (string, error) {
	const baseURL = "https://cloud-api.yandex.net/v1/disk/resources/download"

	params := url.Values{}
	params.Set("path", path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "OAuth "+d.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get download URL")
	}

	var downloadURL struct {
		Href string `json:"href"`
	}

	err = json.NewDecoder(resp.Body).Decode(&downloadURL)
	if err != nil {
		return "", err
	}

	return downloadURL.Href, nil
}
