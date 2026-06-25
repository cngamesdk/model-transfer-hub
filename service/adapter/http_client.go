package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/model"
	"io"
	"net/http"
	"time"
)

// sendChatCompletionRequest 发送聊天完成请求
func sendChatCompletionRequest(ctx context.Context, baseURL, apiKey string, timeout int, req *model.ChatCompletionRequest) (*model.ChatCompletionResponse, any, error) {
	url := fmt.Sprintf("%s/chat/completions", baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result model.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, &result, nil
}

// sendChatCompletionStreamRequest 发送流式聊天完成请求
func sendChatCompletionStreamRequest(ctx context.Context, baseURL, apiKey string, _ int, req *model.ChatCompletionRequest) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/chat/completions", baseURL)

	// 确保stream为true
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("Accept", "text/event-stream")

	// 流式请求不设置超时，避免长时间流式响应被中断
	client := &http.Client{
		Timeout: 0,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}

// sendCompletionRequest 发送文本完成请求
func sendCompletionRequest(ctx context.Context, baseURL, apiKey string, timeout int, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	url := fmt.Sprintf("%s/completions", baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result model.CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// sendCompletionStreamRequest 发送流式文本完成请求
func sendCompletionStreamRequest(ctx context.Context, baseURL, apiKey string, _ int, req *model.CompletionRequest) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/completions", baseURL)

	// 确保stream为true
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("Accept", "text/event-stream")

	// 流式请求不设置超时，避免长时间流式响应被中断
	client := &http.Client{
		Timeout: 0,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
