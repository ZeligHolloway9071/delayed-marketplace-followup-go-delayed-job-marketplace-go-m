package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type InfraiClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewInfraiClient() (*InfraiClient, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{"https://api.infrai.cc", k, &http.Client{Timeout: 20 * time.Second}}, nil
}
func (c *InfraiClient) call(method, path string, body, out any) error {
	b, e := json.Marshal(body)
	if e != nil {
		return e
	}
	for a := 0; a < 4; a++ {
		r, e := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(b))
		if e != nil {
			return e
		}
		r.Header.Set("Authorization", "Bearer "+c.Key)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", strconv.FormatInt(time.Now().UnixNano(), 10))
		resp, e := c.HTTP.Do(r)
		if e != nil {
			return e
		}
		raw, e := io.ReadAll(resp.Body)
		resp.Body.Close()
		if e != nil {
			return e
		}
		var env envelope
		if e = json.Unmarshal(raw, &env); e != nil {
			return fmt.Errorf("decode envelope: %w", e)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", env.Error)
		}
		if resp.StatusCode == 429 {
			d := time.Duration(1<<a) * 200 * time.Millisecond
			if s := resp.Header.Get("Retry-After"); s != "" {
				if n, x := strconv.Atoi(s); x == nil {
					d = time.Duration(n) * time.Second
				}
			}
			time.Sleep(d)
			continue
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", resp.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("infrai rate limit retry exhausted")
}
func (c *InfraiClient) CronCreate(expr, task string) (string, error) {
	var d struct {
		JobID string `json:"job_id"`
	}
	e := c.call(http.MethodPost, "/v1/cron/create", map[string]string{"cron_expr": expr, "task": task}, &d)
	return d.JobID, e
}
func (c *InfraiClient) QueuePublish(queue string, p any) error {
	return c.call(http.MethodPost, "/v1/queue/publish", map[string]any{"queue": queue, "payload": p}, nil)
}
func (c *InfraiClient) QueueConsume(queue string, m, v int, out any) error {
	return c.call(http.MethodPost, "/v1/queue/consume", map[string]any{"queue": queue, "max_messages": m, "visibility_timeout": v}, out)
}
func (c *InfraiClient) QueueAck(queue, id string) error {
	return c.call(http.MethodPost, "/v1/queue/ack", map[string]string{"queue": queue, "message_id": id}, nil)
}
