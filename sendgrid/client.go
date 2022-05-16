package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/go-kit/log"
)

const (
	endpoint   = "https://api.sendgrid.com/v3/"
	apiTimeout = 10 * time.Second
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	logger     log.Logger
}

func NewDefaultClient(apiKey string, logger log.Logger) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		logger:     logger,
	}
}

type requestParams struct {
	method       string
	subPath      string
	queries      map[string]string
	arrayQueries map[string][]string
	body         io.Reader
}

func (c *Client) doAPIRequest(ctx context.Context, params *requestParams, out interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)

	defer cancel()

	req, err := c.newRequest(ctx, params)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Log("error body: %s\n", string(body))

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.Unmarshal(body, out)

	return err
}

func (c *Client) newRequest(ctx context.Context, params *requestParams) (*http.Request, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(url.Path, params.subPath)

	req, err := http.NewRequestWithContext(ctx, params.method, url.String(), params.body)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()

	for k, v := range params.queries {
		query.Add(k, v)
	}

	for k, vs := range params.arrayQueries {
		for _, v := range vs {
			query.Add(k, v)
		}
	}

	req.URL.RawQuery = query.Encode()

	return req, nil
}
