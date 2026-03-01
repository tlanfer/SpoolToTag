package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tlanfer/SpoolToTag/openspool"
)

type Analyzer interface {
	Analyze(ctx context.Context, imageData []byte, contentType string) (openspool.SpoolData, error)
}

type Client struct {
	APIKey  string
	Model   string
	BaseURL string
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: "https://api.openai.com",
	}
}

func (c *Client) Analyze(ctx context.Context, imageData []byte, contentType string) (openspool.SpoolData, error) {
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(imageData))

	falseVal := false
	req := chatRequest{
		Model: c.Model,
		Messages: []message{
			{
				Role: "user",
				Content: []content{
					{
						Type: "text",
						Text: "Extract the filament spool information from this label image. Return the filament type (e.g. PLA, PETG, ABS), the color as a hex code, the brand name, and the recommended min and max nozzle temperatures in Celsius.",
					},
					{
						Type:     "image_url",
						ImageURL: &imageURL{URL: dataURL},
					},
				},
			},
		},
		ResponseFormat: responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchema{
				Name:   "filament_info",
				Strict: true,
				Schema: schemaEntry{
					Type: "object",
					Properties: map[string]schemaEntry{
						"type":      {Type: "string", Description: "Filament type, e.g. PLA, PETG, ABS, TPU"},
						"color_hex": {Type: "string", Description: "Single primary color as one hex code, e.g. #FF5733. Only return one color."},
						"brand":     {Type: "string", Description: "Brand name. Must be one of: Generic, Overture, PolyLite, eSun, PolyTerra"},
						"min_temp":  {Type: "integer", Description: "Minimum nozzle temperature in Celsius"},
						"max_temp":  {Type: "integer", Description: "Maximum nozzle temperature in Celsius"},
					},
					Required:             []string{"type", "color_hex", "brand", "min_temp", "max_temp"},
					AdditionalProperties: &falseVal,
				},
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return openspool.SpoolData{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return openspool.SpoolData{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return openspool.SpoolData{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openspool.SpoolData{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return openspool.SpoolData{}, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, respBody)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return openspool.SpoolData{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return openspool.SpoolData{}, fmt.Errorf("no choices in response")
	}

	var extracted extractedData
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &extracted); err != nil {
		return openspool.SpoolData{}, fmt.Errorf("unmarshal extracted data: %w", err)
	}

	colorHex := strings.Split(extracted.ColorHex, ",")[0]
	colorHex = strings.TrimSpace(colorHex)

	brand := openspool.NormalizeBrand(extracted.Brand)

	return openspool.New(extracted.Type, colorHex, brand, extracted.MinTemp, extracted.MaxTemp)
}
