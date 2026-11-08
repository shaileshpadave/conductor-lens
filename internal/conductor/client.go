package conductor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetWorkflow(workflowID string) (*Workflow, error) {
	endpoint := fmt.Sprintf("%s/api/workflow/%s?includeTasks=true", c.baseURL, workflowID)
	var wf Workflow
	if err := c.get(endpoint, &wf); err != nil {
		return nil, fmt.Errorf("fetching workflow %s: %w", workflowID, err)
	}
	return &wf, nil
}

func (c *Client) SearchFailed(workflowName string, limit int) (*SearchResult, error) {
	query := "status:FAILED"
	if workflowName != "" {
		query += fmt.Sprintf(" AND workflowType:%s", workflowName)
	}
	endpoint := fmt.Sprintf("%s/api/workflow/search?query=%s&size=%d&sort=startTime:DESC",
		c.baseURL, url.QueryEscape(query), limit)

	var result SearchResult
	if err := c.get(endpoint, &result); err != nil {
		return nil, fmt.Errorf("searching workflows: %w", err)
	}
	return &result, nil
}

func (c *Client) get(endpoint string, out any) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Authorization", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("conductor returned %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, out)
}
