package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL   = "https://howlongtobeat.com"
	searchAPI = baseURL + "/api/search/site"
	initAPI   = searchAPI + "/init"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36"
)

var (
	errTooManyRequests = errors.New("rate limited by HowLongToBeat, wait a minute and retry")
	errTokenRejected   = errors.New("search token rejected")
)

// Game holds the subset of HLTB fields the CLI uses. Times are in seconds.
type Game struct {
	ID          int    `json:"game_id"`
	Name        string `json:"game_name"`
	Alias       string `json:"game_alias"`
	Type        string `json:"game_type"`
	Image       string `json:"game_image"`
	Main        int    `json:"comp_main"`
	MainExtra   int    `json:"comp_plus"`
	Completion  int    `json:"comp_100"`
	AllStyles   int    `json:"comp_all"`
	MainCount   int    `json:"comp_main_count"`
	ExtraCount  int    `json:"comp_plus_count"`
	CompCount   int    `json:"comp_100_count"`
	AllCount    int    `json:"comp_all_count"`
	CoOp        int    `json:"invested_co"`
	Multi       int    `json:"invested_mp"`
	ReviewScore int    `json:"review_score"`
	Playing     int    `json:"count_playing"`
	Backlog     int    `json:"count_backlog"`
	Completed   int    `json:"count_comp"`
	Platforms   string `json:"profile_platform"`
	ReleaseYear int    `json:"release_world"`
}

// SearchResult is the paginated response of the search endpoint.
type SearchResult struct {
	Count       int    `json:"count"`
	PageCurrent int    `json:"pageCurrent"`
	PageTotal   int    `json:"pageTotal"`
	Data        []Game `json:"data"`
}

// URL returns the public HLTB page of the game.
func (g Game) URL() string {
	return fmt.Sprintf("%s/game/%d", baseURL, g.ID)
}

// Client talks to the unofficial HLTB search API.
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 20 * time.Second}}
}

// Search queries HLTB; it refreshes the auth token once if the first one is rejected.
func (c *Client) Search(ctx context.Context, query string, page, size int) (*SearchResult, error) {
	token, err := c.fetchToken(ctx)
	if err != nil {
		return nil, err
	}

	res, err := c.search(ctx, token, query, page, size)
	if !errors.Is(err, errTokenRejected) {
		return res, err
	}

	if token, err = c.fetchToken(ctx); err != nil {
		return nil, err
	}
	return c.search(ctx, token, query, page, size)
}

func (c *Client) fetchToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s?t=%d", initAPI, time.Now().UnixMilli())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	setCommonHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching token: unexpected status %s", resp.Status)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || payload.Token == "" {
		return "", errors.New("fetching token: unexpected response (the HLTB API may have changed)")
	}
	return payload.Token, nil
}

func (c *Client) search(ctx context.Context, token, query string, page, size int) (*SearchResult, error) {
	body, err := json.Marshal(buildPayload(query, page, size))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchAPI, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", baseURL)
	req.Header.Set("x-auth-token", token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusForbidden:
		return nil, errTokenRejected
	case http.StatusTooManyRequests:
		return nil, errTooManyRequests
	default:
		return nil, fmt.Errorf("searching: unexpected status %s", resp.Status)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	var res SearchResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("decoding response (the HLTB API may have changed): %w", err)
	}
	return &res, nil
}

func setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", baseURL+"/")
	req.Header.Set("Accept", "*/*")
}

func buildPayload(query string, page, size int) map[string]any {
	return map[string]any{
		"searchType":  "games",
		"searchTerms": strings.Fields(query),
		"searchPage":  page,
		"size":        size,
		"searchOptions": map[string]any{
			"games": map[string]any{
				"userId":        0,
				"platform":      "",
				"sortCategory":  "popular",
				"rangeCategory": "main",
				"rangeTime":     map[string]int{"min": 0, "max": 0},
				"gameplay":      map[string]string{"perspective": "", "flow": "", "genre": ""},
				"modifier":      "",
			},
			"users":      map[string]string{"sortCategory": "postcount"},
			"lists":      map[string]string{"sortCategory": "follows"},
			"filter":     "",
			"sort":       0,
			"randomizer": 0,
		},
		"useCache": true,
	}
}
