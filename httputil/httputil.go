package httputil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Fetch is an utility funciton to perform a HTTP request
func Fetch(method string, url string, body []byte, headers map[string]string) ([]byte, error) {
	payload := bytes.NewReader(body)
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// FetchJSON is the same as Fetch plus decodes the response as JSON
func FetchJSON(method string, url string, body []byte, headers map[string]string, out interface{}) error {
	response, err := Fetch(method, url, body, headers)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(response, out); err != nil {
		return err
	}

	return nil
}
