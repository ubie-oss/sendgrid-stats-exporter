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

	"github.com/go-kit/log"
)

const (
	endpoint = "https://api.sendgrid.com/v3/"
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
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, params.subPath)

	req, err := http.NewRequest(params.method, u.String(), params.body)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range params.queries {
		q.Add(k, v)
	}

	for k, vs := range params.arrayQueries {
		for _, v := range vs {
			q.Add(k, v)
		}
	}

	req.URL.RawQuery = q.Encode()

	return req, nil
}
