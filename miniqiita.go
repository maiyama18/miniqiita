package miniqiita

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	Token      string
	Logger     *log.Logger
}

func New(rawBaseURL, token string, logger *log.Logger) (*Client, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = log.New(os.Stderr, "[LOG]", log.LstdFlags)
	}

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: http.DefaultClient,
		Token:      token,
		Logger:     logger,
	}, nil
}

type Item struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	LikesCount int    `json:"likes_count"`
}

func (c *Client) GetUserItems(ctx context.Context, userID string, page, perPage int) ([]*Item, error) {
	relativePath := path.Join("users", userID, "items")
	queries := map[string]string{
		"page":     strconv.Itoa(page),
		"per_page": strconv.Itoa(perPage),
	}
	req, err := c.newRequest(ctx, http.MethodGet, relativePath, queries, nil, nil)
	if err != nil {
		return nil, err
	}

	// send request
	var items []*Item
	code, err := c.doRequest(req, &items)

	switch code {
	case http.StatusOK:
		return items, nil
	case http.StatusBadRequest:
		return nil, errors.New("bad request. some parameters may be invalid")
	case http.StatusNotFound:
		return nil, fmt.Errorf("not found. user with id '%s' may not exist", userID)
	default:
		return nil, errors.New("unexpected error")
	}
}

func (c *Client) newRequest(ctx context.Context, method, relativePath string, queries, headers map[string]string, reqBody io.Reader) (*http.Request, error) {
	reqURL := *c.BaseURL

	// set path
	reqURL.Path = path.Join(reqURL.Path, relativePath)

	// set query
	if queries != nil {
		q := reqURL.Query()
		for k, v := range queries {
			q.Add(k, v)
		}
		reqURL.RawQuery = q.Encode()
	}

	// instantiate request
	req, err := http.NewRequest(method, reqURL.String(), reqBody)
	if err != nil {
		return nil, err
	}

	// set header
	req.Header.Set("User-Agent", "qiita-go-client")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// set context
	req = req.WithContext(ctx)

	return req, nil
}

func (c *Client) doRequest(req *http.Request, respBody interface{}) (int, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || 300 <= resp.StatusCode {
		return resp.StatusCode, nil
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if err := json.Unmarshal(bodyBytes, respBody); err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
