package shardkey

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/pkg/logger"
)

type LLM struct {
	APIKey string
	Model  string
}

func NewLLM(apiKey string, model string) *LLM {
	return &LLM{
		APIKey: apiKey,
		Model:  model,
	}
}

func (l *LLM) CallLLM(ctx context.Context, prompt string) (string, error) {

	url := "http://192.168.112.1:11434/api/generate"

	body := map[string]interface{}{
		"model":  l.Model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %s", string(body))
	}

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Response == "" {
		return "", errors.New("empty Ollama response")
	}

	return result.Response, nil
}

func (s *InferenceService) RunLLMInference(
	ctx context.Context,
	projectID string,
	config *repository.AIConfig,
) error {

	logger.Logger.Info("LLM inference started")

	// Build schema
	logicalSchema, err := s.buildSchema(ctx, projectID)
	if err != nil {
		return err
	}

	// Convert to LLM-safe schema
	llmSchema := buildLLMSchema(logicalSchema)

	schemaJSON, err := json.MarshalIndent(llmSchema, "", "  ")
	if err != nil {
		return err
	}

	// Build prompt
	prompt := buildPrompt(schemaJSON)

	// Call LLM
	llm := NewLLM(config.APIKey, config.Model)

	responseText, err := llm.CallLLM(ctx, prompt)
	if err != nil {
		return err
	}

	fmt.Println("RAW LLM RESPONSE:\n", responseText)

	start := strings.Index(responseText, "{")
	if start != -1 {
		responseText = responseText[start:]
	}

	var parsed struct {
		Decisions []struct {
			Table  string `json:"table"`
			Column string `json:"column"`
		} `json:"decisions"`
	}

	if err := json.Unmarshal([]byte(responseText), &parsed); err != nil {
		return fmt.Errorf("failed to parse LLM response: %v\nResponse: %s", err, responseText)
	}

	if len(parsed.Decisions) == 0 {
		return errors.New("LLM returned no decisions")
	}

	records := make([]repository.ShardKeyRecord, 0, len(parsed.Decisions))

	for _, d := range parsed.Decisions {
		records = append(records, repository.ShardKeyRecord{
			TableName:      d.Table,
			ShardKeyColumn: d.Column,
			IsManual:       false,
		})
	}

	return s.shardKeyRepo.ReplaceShardKeysForProject(ctx, projectID, records)
}
