package server_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"geolineai/internal/core/assist"
	"geolineai/internal/core/assist/prompts"
	"geolineai/internal/core/intake/importer"
	"geolineai/internal/core/scenarios"
	"geolineai/internal/integrations/db"
	"geolineai/internal/server"
)

var (
	harnessPort   = flag.Int("harness-port", 0, "serve the browser harness on this port until interrupted; 0 skips it")
	harnessModels = flag.String("harness-models", "scripted", "scripted: answers chosen by the question; none: no provider key")
)

// scripted answers by the question's keywords, so a browser check can reach every
// outcome without a provider: "withheld", "unanswerable", "invalid", "provider",
// "slow" (a shown answer after three seconds) and anything else (shown).
type scripted struct{}

func (scripted) Write(ctx context.Context, _, user string, _ map[string]any) (json.RawMessage, string, error) {
	var in struct {
		Question string `json:"question"`
		Sources  []struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(user), &in); err != nil {
		return nil, "", err
	}
	q := strings.ToLower(in.Question)
	cite := in.Sources[0]
	switch {
	case strings.Contains(q, "withheld"):
		return json.RawMessage(fmt.Sprintf(`{"answerable":true,"sentences":[{"text":%q,"sources":[%q]},{"text":"It holds 777 readings.","sources":[%q]}]}`, cite.Text, cite.ID, cite.ID)), "scripted-writer", nil
	case strings.Contains(q, "unanswerable"):
		return json.RawMessage(`{"answerable":false,"sentences":[]}`), "scripted-writer", nil
	case strings.Contains(q, "invalid"):
		return json.RawMessage(`{"sentences":[]}`), "scripted-writer", nil
	case strings.Contains(q, "provider"):
		return nil, "", fmt.Errorf("%w: the scripted provider refused", assist.ErrUnavailable)
	case strings.Contains(q, "slow"):
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil, "", ctx.Err()
		}
	}
	return json.RawMessage(fmt.Sprintf(`{"answerable":true,"sentences":[{"text":%q,"sources":[%q]}]}`, cite.Text, cite.ID)), "scripted-writer", nil
}

func (scripted) Judge(_ context.Context, _ any, questions map[string]assist.Question) (map[string]assist.Answer, string, error) {
	return map[string]assist.Answer{"supported": {Kind: assist.KindNoul, Noul: 0.95}, "verdict": {Kind: assist.KindNoul, Noul: 0.02}}, "scripted-judge", nil
}

// The browser checks drive the real server, routes and database with models scripted
// here, never in the shipped binary. Give it a port to drive it from a browser:
//
//	go test ./internal/server -run BrowserHarness -harness-port 18999
func TestBrowserHarness(t *testing.T) {
	if *harnessPort == 0 {
		t.Skip("the browser harness runs only with -harness-port")
	}
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	csv, err := os.ReadFile(repoPath("data", "line-1.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := importer.Import(ctx, d, "/data/line-1.csv", strings.NewReader(string(csv))); err != nil {
		t.Fatal(err)
	}
	manifest, err := scenarios.Load(repoPath("data", "scenarios.json"))
	if err != nil {
		t.Fatal(err)
	}
	var models server.Models
	if *harnessModels == "scripted" {
		models = server.Models{Writer: scripted{}, Judge: scripted{}}
	}
	gin.SetMode(gin.ReleaseMode)
	routes := server.Routes(d, models, server.Config{Manifest: manifest, WebRoot: repoPath("frontend", "dist"), Fingerprint: prompts.Fingerprint})
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(*harnessPort)))
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: routes, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(listener) }()
	fmt.Printf("harness: serving on http://%s with %s models\n", listener.Addr(), *harnessModels)
	stop, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	<-stop.Done()
	_ = srv.Close()
}
