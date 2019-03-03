package exporter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	log "github.com/sirupsen/logrus"
)

// gatherData - Collects the data from the API and stores into struct
func (e *Exporter) gatherData() ([]*github.Repository, *RateLimits, error) {

	data := []*github.Repository{}

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: e.APIToken},
	)

	tokenClient := oauth2.NewClient(context.Background(), tokenSource)

	client := github.NewClient(tokenClient)

	data, err := queryGitHub(client, e.TargetRepos)

	if err != nil {
		return data, nil, err
	}

	// for _, response := range responses {

	// 	// Github can at times present an array, or an object for the same data set.
	// 	// This code checks handles this variation.
	// 	if isArray(response.body) {
	// 		ds := []*github.Repository{}
	// 		json.Unmarshal(response.body, &ds)
	// 		data = append(data, ds...)
	// 	} else {
	// 		d := new(github.Repository)
	// 		json.Unmarshal(response.body, &d)
	// 		data = append(data, d)
	// 	}

	// 	log.Infof("API data fetched for repository: %s", response.url)
	// }

	// Check the API rate data and store as a metric
	rates, err := getRates(e.APIURL, e.APIToken)

	if err != nil {
		log.Errorf("Unable to obtain rate limit data from API, Error: %s", err)
	}

	//return data, rates, err
	return data, rates, nil

}

func queryGitHub(client *github.Client, repoMap map[string][]string) ([]*github.Repository, error) {
	infos := []*github.Repository{}

	for owner, repoList := range repoMap {

		for _, repo := range repoList {
			if repo == "*" {
				listRepoOptions := &github.RepositoryListOptions{
					ListOptions: github.ListOptions{
						PerPage: 100,
						Page:    0,
					},
				}

				repoDatas, response, err := client.Repositories.List(context.Background(), owner, listRepoOptions)

				if err != nil {
					return infos, err
				}
				if response.StatusCode != http.StatusOK {
					return infos, fmt.Errorf("HTTP status unexpected: %d for %s/%s", response.StatusCode, owner, repo)
				}
				infos = append(infos, repoDatas...)
			}
		}
	}

	return infos, nil
}

// getRates obtains the rate limit data for requests against the github API.
// Especially useful when operating without oauth and the subsequent lower cap.
func getRates(baseURL string, token string) (*RateLimits, error) {

	rateEndPoint := ("/rate_limit")
	url := fmt.Sprintf("%s%s", baseURL, rateEndPoint)

	resp, err := getHTTPResponse(url, token)
	if err != nil {
		return &RateLimits{}, err
	}
	defer resp.Body.Close()

	// Triggers if rate-limiting isn't enabled on private Github Enterprise installations
	if resp.StatusCode == 404 {
		return &RateLimits{}, fmt.Errorf("Rate Limiting not enabled in GitHub API")
	}

	limit, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Limit"), 64)

	if err != nil {
		return &RateLimits{}, err
	}

	rem, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Remaining"), 64)

	if err != nil {
		return &RateLimits{}, err
	}

	reset, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Reset"), 64)

	if err != nil {
		return &RateLimits{}, err
	}

	return &RateLimits{
		Limit:     limit,
		Remaining: rem,
		Reset:     reset,
	}, err

}

// isArray simply looks for key details that determine if the JSON response is an array or not.
func isArray(body []byte) bool {

	isArray := false

	for _, c := range body {
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			continue
		}
		isArray = c == '['
		break
	}

	return isArray

}
