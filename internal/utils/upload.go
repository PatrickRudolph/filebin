package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PatrickRudolph/filebin/internal/filedata"
)

type HTTPFileUploader struct {
	Url      string
	Username string
	Password string
}

func (h *HTTPFileUploader) Delete(id string) (response string, err error) {

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodDelete, h.Url+"/"+id, nil)
	if err != nil {
		return
	}
	if h.Username != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}
	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if !strings.Contains(resp.Status, "200") && !strings.Contains(resp.Status, "OK") {
		err = fmt.Errorf("Got HTTP Status %s", resp.Status)
		return
	}

	response = string(respBody)
	return
}

func (h *HTTPFileUploader) Upload(src string, name string) (url string, err error) {
	// Create client
	client := &http.Client{}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("file", name)
	if err != nil {
		err = fmt.Errorf("Failed to create form file: %v", err)
		return
	}

	// open file handle
	fh, err := os.Open(src)
	if err != nil {
		err = fmt.Errorf("Failed to open file: %v", err)
		return
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		err = fmt.Errorf("I/O error: %v", err)
		return
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	req, err := http.NewRequest(http.MethodPost, h.Url, bodyBuf)
	if h.Username != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}
	req.Header.Add("Content-Type", contentType)
	resp, err := client.Do(req)

	if err != nil {
		err = fmt.Errorf("HTTP post failed with: %v", err)
		return
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("I/O Read error: %v", err)
		return
	}

	if !strings.Contains(resp.Status, "200") && !strings.Contains(resp.Status, "OK") {
		err = fmt.Errorf("Got HTTP Status %s", resp.Status)
		return
	}

	url = string(resp_body)
	if !strings.Contains(url, "http://") && !strings.Contains(url, "https://") {
		url = h.Url + "/" + url
	}
	return
}

func (h *HTTPFileUploader) WaitForEvent() (err error) {

	// Create client
	client := &http.Client{
		Timeout: time.Minute * 10,
	}

	// Create request
	req, err := http.NewRequest(http.MethodGet, h.Url+"/event", nil)
	if err != nil {
		return
	}
	if h.Username != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}
	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read Response Body
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if !strings.Contains(resp.Status, "200") && !strings.Contains(resp.Status, "OK") {
		err = fmt.Errorf("Got HTTP Status %s", resp.Status)
		return
	}
	return
}

func (h *HTTPFileUploader) List() (fds []filedata.FileData, err error) {

	// Create client
	client := &http.Client{
		Timeout: time.Minute * 10,
	}

	// Create request
	req, err := http.NewRequest(http.MethodGet, h.Url+"/list", nil)
	if err != nil {
		return
	}
	if h.Username != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read Response Body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if !strings.Contains(resp.Status, "200") && !strings.Contains(resp.Status, "OK") {
		err = fmt.Errorf("Got HTTP Status %s", resp.Status)
		return
	}
	err = json.Unmarshal(body, &fds)

	return
}
