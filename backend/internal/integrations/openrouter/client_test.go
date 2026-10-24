package openrouter_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/core/assist"
	"geolineai/internal/integrations/openrouter"
)

// provider is a local stand-in for both endpoints: it records every call and answers
// with the next scripted status and the current body for the endpoint called.
type provider struct {
	calls  []call
	status []int
	judge  string
	chat   string
}

type call struct {
	path, auth string
	body       map[string]any
}

func (p *provider) serve(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)
	p.calls = append(p.calls, call{r.URL.Path, r.Header.Get("Authorization"), body})
	code := p.status[0]
	if len(p.status) > 1 {
		p.status = p.status[1:]
	}
	w.WriteHeader(code)
	if r.URL.Path == "/one" {
		_, _ = w.Write([]byte(p.judge))
	} else {
		_, _ = w.Write([]byte(p.chat))
	}
}

func newClient(t *testing.T, p *provider) *openrouter.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(p.serve))
	t.Cleanup(srv.Close)
	c := openrouter.New("secret", openrouter.Endpoints{SystemOne: srv.URL + "/one", Chat: srv.URL + "/chat"})
	c.RetryWait = 0
	return c
}

var questions = map[string]assist.Question{
	"supported": {Kind: assist.KindNoul, Instructions: "supported?", Criteria: map[string]string{"yes": "y"}},
	"verdict":   {Kind: assist.KindNoul, Instructions: "verdict?", Criteria: map[string]string{"yes": "y"}},
}

const goodJudgment = `{"model":"typesafe/jev-1.13-x","answers":{"supported":{"type":"noul","noul":0.92},"verdict":{"type":"noul","noul":0.04}}}`

// A wire shape the provider does not expect would come back as near-zero nouls or an
// empty completion and withhold every answer, so the request each endpoint receives
// and the answers decoded are pinned here.
func TestClientSpeaksBothEndpoints(t *testing.T) {
	p := &provider{status: []int{http.StatusOK}, judge: goodJudgment,
		chat: `{"model":"anthropic/x","choices":[{"message":{"content":"Here:\n{\"answerable\":false,\"sentences\":[]}\nDone."}}]}`}
	c := newClient(t, p)
	ctx := context.Background()

	answers, model, err := c.Judge(ctx, map[string]string{"sentence": "s"}, questions)
	require.NoError(t, err)
	assert.Equal(t, "typesafe/jev-1.13-x", model)
	assert.Equal(t, map[string]assist.Answer{
		"supported": {Kind: assist.KindNoul, Noul: 0.92},
		"verdict":   {Kind: assist.KindNoul, Noul: 0.04},
	}, answers)
	assert.Equal(t, call{"/one", "Bearer secret", map[string]any{
		"model": "typesafe/jev-1.13",
		"state": map[string]any{"sentence": "s"},
		"questions": map[string]any{
			"supported": map[string]any{"type": "noul", "instructions": "supported?", "criteria": map[string]any{"yes": "y"}},
			"verdict":   map[string]any{"type": "noul", "instructions": "verdict?", "criteria": map[string]any{"yes": "y"}},
		},
	}}, p.calls[0])

	object, model, err := c.Write(ctx, "the system", "the user", map[string]any{"type": "object"})
	require.NoError(t, err)
	assert.Equal(t, "anthropic/x", model)
	assert.JSONEq(t, `{"answerable":false,"sentences":[]}`, string(object), "the object alone, whatever surrounds it")
	assert.Equal(t, "/chat", p.calls[1].path)
	assert.Equal(t, "anthropic/claude-sonnet-5.5", p.calls[1].body["model"])
	assert.Equal(t, float64(0), p.calls[1].body["temperature"])
	assert.Equal(t, []any{map[string]any{"role": "system", "content": "the system"}, map[string]any{"role": "user", "content": "the user"}}, p.calls[1].body["messages"])
	assert.Equal(t, map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "answer", "strict": true, "schema": map[string]any{"type": "object"}}}, p.calls[1].body["response_format"])
}

// A rate-limited or overloaded answer is retried once, as configured; one still
// refused after that, a refused key and an unreachable host are provider errors.
func TestClientRetriesThenFails(t *testing.T) {
	ctx := context.Background()
	p := &provider{status: []int{http.StatusTooManyRequests, http.StatusOK}, judge: goodJudgment}
	c := newClient(t, p)
	_, _, err := c.Judge(ctx, "s", questions)
	require.NoError(t, err)
	assert.Len(t, p.calls, 2, "one retry")

	for _, tc := range []struct {
		name  string
		codes []int
		calls int
	}{
		{"overloaded twice", []int{529, 529}, 2},
		{"rate-limited twice", []int{http.StatusTooManyRequests, http.StatusTooManyRequests}, 2},
		{"refused key, not retried", []int{http.StatusUnauthorized}, 1},
	} {
		p := &provider{status: tc.codes, judge: goodJudgment}
		_, _, err := newClient(t, p).Judge(ctx, "s", questions)
		assert.ErrorIs(t, err, assist.ErrUnavailable, tc.name)
		assert.NotErrorIs(t, err, assist.ErrInvalidResponse, tc.name)
		assert.Len(t, p.calls, tc.calls, tc.name)
	}
	down := openrouter.New("secret", openrouter.Endpoints{SystemOne: "http://127.0.0.1:1/one", Chat: "http://127.0.0.1:1/chat"})
	_, _, err = down.Write(ctx, "s", "u", nil)
	assert.ErrorIs(t, err, assist.ErrUnavailable)
}

// An answer that cannot be read is an invalid response, never a provider error and
// never an answer: a missing or out-of-range probability would otherwise pass or
// fail every threshold. Exactly 0 and exactly 1 are probabilities like any other.
func TestClientRefusesUnreadableAnswers(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct{ name, judge string }{
		{"malformed JSON", `{"answers":`},
		{"a question unanswered", `{"answers":{"supported":{"type":"noul","noul":0.9}}}`},
		{"another kind of answer", `{"answers":{"supported":{"type":"choice","noul":0.9},"verdict":{"type":"noul","noul":0.1}}}`},
		{"no probability", `{"answers":{"supported":{"type":"noul"},"verdict":{"type":"noul","noul":0.1}}}`},
		{"a probability as text", `{"answers":{"supported":{"type":"noul","noul":"0.9"},"verdict":{"type":"noul","noul":0.1}}}`},
		{"below 0", `{"answers":{"supported":{"type":"noul","noul":-0.2},"verdict":{"type":"noul","noul":0.1}}}`},
		{"above 1", `{"answers":{"supported":{"type":"noul","noul":1.7},"verdict":{"type":"noul","noul":0.1}}}`},
	} {
		_, _, err := newClient(t, &provider{status: []int{http.StatusOK}, judge: tc.judge}).Judge(ctx, "s", questions)
		assert.ErrorIs(t, err, assist.ErrInvalidResponse, tc.name)
		assert.NotErrorIs(t, err, assist.ErrUnavailable, tc.name)
	}
	answers, _, err := newClient(t, &provider{status: []int{http.StatusOK},
		judge: `{"answers":{"supported":{"type":"noul","noul":1},"verdict":{"type":"noul","noul":0}}}`}).Judge(ctx, "s", questions)
	require.NoError(t, err)
	assert.Equal(t, 1.0, answers["supported"].Noul)
	assert.Equal(t, 0.0, answers["verdict"].Noul)

	for _, tc := range []struct{ name, chat string }{
		{"malformed completion", `{"choices":`},
		{"no choice", `{"choices":[]}`},
		{"no JSON object", `{"choices":[{"message":{"content":"I cannot answer."}}]}`},
		{"a broken object", `{"choices":[{"message":{"content":"{\"answerable\": tru"}}]}`},
	} {
		_, _, err := newClient(t, &provider{status: []int{http.StatusOK}, chat: tc.chat}).Write(ctx, "s", "u", nil)
		assert.ErrorIs(t, err, assist.ErrInvalidResponse, tc.name)
	}
}
