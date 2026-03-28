package shardkey

import "context"

func PingLLM(ctx context.Context, apiKey string, model string) error {
	llm := NewLLM(apiKey, model)
	_, err := llm.CallLLM(ctx, "Reply with OK")
	return err
}
