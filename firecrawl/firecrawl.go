package firecrawl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	apiKey     string
	apiUrl     string
	httpClient *http.Client
}

type DocumentMetadata struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type ExtractionData struct {
	CompanyDescription string `json:"company_description,omitempty"`
	ProductSummary     string `json:"product_summary,omitempty"`
}

type ResponseData struct {
	Metadata *DocumentMetadata `json:"metadata,omitempty"`
	Extract  *ExtractionData   `json:"extract,omitempty"`
}

type ScrapeResponse struct {
	Success bool          `json:"success"`
	Data    *ResponseData `json:"data,omitempty"`
}

func NewClient() *Client {
	return &Client{
		apiKey:     os.Getenv("FIRECRAWL_API_KEY"),
		apiUrl:     os.Getenv("FIRECRAWL_API_URL"),
		httpClient: &http.Client{},
	}
}

func (c *Client) ScrapeURL(url string) (*ScrapeResponse, error) {
	params := map[string]any{
		"url":     url,
		"formats": []string{"extract"},
		"extract": map[string]any{
			"prompt": "Extract a company description and summary of its products from this page.",
			"schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"company_description": map[string]any{
						"type": "string",
					},
					"product_summary": map[string]any{
						"type": "string",
					},
				},
				"required": []string{
					"company_description",
					"product_summary",
				},
			},
		},
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", c.apiKey),
	}

	reqUrl := fmt.Sprintf("%s/v1/scrape", c.apiUrl)
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var scrapeResponse ScrapeResponse
	if err = json.Unmarshal(respBody, &scrapeResponse); err != nil {
		return nil, err
	}

	if !scrapeResponse.Success {
		return nil, fmt.Errorf("Scrape failed")
	}

	return &scrapeResponse, nil
}
