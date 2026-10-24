package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"geolineai/internal/core/assist"
	"geolineai/internal/core/assist/prompts"
)

// Endpoints are the two URLs the client posts to.
type Endpoints struct {
	SystemOne string
	Chat      string
}

// DefaultEndpoints are OpenRouter's own.
var DefaultEndpoints = Endpoints{
	SystemOne: "https://openrouter.ai/api/v1/systemone",
	Chat:      "https://openrouter.ai/api/v1/chat/completions",
}

// The models, pinned: the thresholds were chosen against these, so a change is part
// of the configuration fingerprint and needs an evaluation review.
var (
	judgeModel  = prompts.Settings.Jev.Model
	writerModel = prompts.Settings.LLM.Model
)

// timeout bounds one call; the page waits on it.
var timeout = time.Duration(prompts.Settings.LLM.TimeoutSeconds) * time.Second

// The retry policy: an answer with one of retryStatuses, a rate-limited or overloaded
// one, is retried up to retries times, each after the client's RetryWait.
var (
	retryStatuses = prompts.Settings.LLM.RetryStatuses
	retries       = prompts.Settings.LLM.Retries
)

// Client is one key over the two endpoints. A request that fails, or is still
// refused after the retries, is an error wrapping the unavailable error; an answer of
// the wrong shape is one wrapping the invalid-response error.
type Client struct {
	key       string
	endpoints Endpoints
	http      *http.Client
	// RetryWait is how long a rate-limited or overloaded answer waits before each
	// retry; a test sets it to zero.
	RetryWait time.Duration
}

// New makes a client for the key at the endpoints.
func New(key string, endpoints Endpoints) *Client {
	return &Client{key: key, endpoints: endpoints, http: &http.Client{Timeout: timeout}, RetryWait: time.Duration(prompts.Settings.LLM.RetryWaitSeconds) * time.Second}
}

// question is a typed question as the System One endpoint reads it.
type question struct {
	Type         string            `json:"type"`
	Instructions any               `json:"instructions"`
	Criteria     map[string]string `json:"criteria,omitempty"`
}

// answer is one answer as the endpoint writes it.
type answer struct {
	Type string   `json:"type"`
	Noul *float64 `json:"noul"`
}

// Judge posts the questions over the state and returns every answer. An answer that
// is missing, of another kind than its question, without a probability, or with one
// outside 0 to 1 is an invalid response: a missing probability would otherwise read
// as zero, and one out of range would pass or fail every threshold.
func (c *Client) Judge(ctx context.Context, state any, questions map[string]assist.Question) (map[string]assist.Answer, string, error) {
	qs := make(map[string]question, len(questions))
	for name, q := range questions {
		qs[name] = question{Type: string(q.Kind), Instructions: q.Instructions, Criteria: q.Criteria}
	}
	body, err := c.post(ctx, c.endpoints.SystemOne, map[string]any{"model": judgeModel, "state": state, "questions": qs})
	if err != nil {
		return nil, "", err
	}
	var res struct {
		Model   string            `json:"model"`
		Answers map[string]answer `json:"answers"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, "", fmt.Errorf("%w: decode the judgment: %v", assist.ErrInvalidResponse, err)
	}
	out := make(map[string]assist.Answer, len(questions))
	for name, q := range questions {
		a, ok := res.Answers[name]
		switch {
		case !ok || a.Type != string(q.Kind):
			return nil, "", fmt.Errorf("%w: the judgment has no %s answer for %q", assist.ErrInvalidResponse, q.Kind, name)
		case a.Noul == nil:
			return nil, "", fmt.Errorf("%w: the noul answer for %q has no probability", assist.ErrInvalidResponse, name)
		case *a.Noul < 0 || *a.Noul > 1:
			return nil, "", fmt.Errorf("%w: the noul answer for %q is %v, outside 0 to 1", assist.ErrInvalidResponse, name, *a.Noul)
		}
		out[name] = assist.Answer{Kind: q.Kind, Noul: *a.Noul}
	}
	return out, res.Model, nil
}

// Write posts one chat completion, the system and user messages as given, held to
// the schema, at temperature zero, and returns the JSON object the model wrote, taken
// from the first brace to the end of that object so a fence or a sentence around it
// does not count.
func (c *Client) Write(ctx context.Context, system, user string, schema map[string]any) (json.RawMessage, string, error) {
	body, err := c.post(ctx, c.endpoints.Chat, map[string]any{
		"model":       writerModel,
		"temperature": prompts.Settings.LLM.Temperature,
		"messages":    []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": user}},
		"response_format": map[string]any{"type": "json_schema", "json_schema": map[string]any{
			"name": "answer", "strict": true, "schema": schema,
		}},
	})
	if err != nil {
		return nil, "", err
	}
	var res struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, "", fmt.Errorf("%w: decode the completion: %v", assist.ErrInvalidResponse, err)
	}
	if len(res.Choices) == 0 {
		return nil, "", fmt.Errorf("%w: the completion holds no choice", assist.ErrInvalidResponse)
	}
	content := res.Choices[0].Message.Content
	start := strings.Index(content, "{")
	if start < 0 {
		return nil, "", fmt.Errorf("%w: the completion holds no JSON object: %.200q", assist.ErrInvalidResponse, content)
	}
	var object json.RawMessage
	if err := json.NewDecoder(strings.NewReader(content[start:])).Decode(&object); err != nil {
		return nil, "", fmt.Errorf("%w: the completion's JSON object does not parse: %v", assist.ErrInvalidResponse, err)
	}
	return object, res.Model, nil
}

// post sends one JSON body with the key and returns the answer's body. An answer with
// a retried status is retried as the retry policy says, after RetryWait; any other
// non-200 answer, an answer still refused after the last retry, and any transport
// failure is an unavailable error naming the status and the start of the body.
func (c *Client) post(ctx context.Context, url string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", assist.ErrUnavailable, err)
		}
		body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		_ = res.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", assist.ErrUnavailable, err)
		}
		switch {
		case res.StatusCode == http.StatusOK:
			return body, nil
		case slices.Contains(retryStatuses, res.StatusCode) && attempt < retries:
			select {
			case <-time.After(c.RetryWait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		default:
			return nil, fmt.Errorf("%w: %s answered %d: %.200s", assist.ErrUnavailable, url, res.StatusCode, body)
		}
	}
}
