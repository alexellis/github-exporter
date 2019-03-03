package exporter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// queryAPI queries the GitHub API asynchronously
func queryAPI(targets []string, token string) ([]*Response, error) {

	// Channels used to enable concurrent requests
	ch := make(chan *Response, len(targets))

	responses := []*Response{}

	for _, url := range targets {

		go func(url string) {
			err := getResponse(url, token, ch)
			if err != nil {
				ch <- &Response{url, nil, []byte{}, err}
			}
		}(url)

	}

	for {
		select {
		case r := <-ch:
			if r.err != nil {
				log.Errorf("Error scraping API, Error: %v", r.err)
				break
			}
			responses = append(responses, r)

			if len(responses) == len(targets) {
				return responses, nil
			}
		}

	}
}

// getResponse collects an individual http.response and returns a *Response
func getResponse(url string, token string, ch chan<- *Response) error {

	log.Infof("Fetching %s \n", url)

	resp, err := getHTTPResponse(url, token) // do this earlier

	if err != nil {
		return fmt.Errorf("Error converting body to byte array: %v", err)
	}

	// Read the body to a byte array so it can be used elsewhere
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("Error converting body to byte array: %v", err)
	}

	defer resp.Body.Close()

	// Triggers if a user specifies an invalid or not visible repository
	if resp.StatusCode == 404 {
		return fmt.Errorf("received 404 status from Github API, ensure the repsository URL is correct, if this is a private repo then check the token scope")
	}

	ch <- &Response{url, resp, body, err}

	return nil
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func getHTTPResponse(url string, token string) (*http.Response, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// If a token is present, add it to the http.request
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, err
}
