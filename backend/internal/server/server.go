package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"geolineai/internal/core/assist"
	"geolineai/internal/core/assist/explain"
	"geolineai/internal/core/ids"
	"geolineai/internal/core/review/readings"
	"geolineai/internal/core/review/verdict"
	"geolineai/internal/core/scenarios"
)

// Store is all the storage the server reads and writes: each core package's own
// needs, and a count of imports, which fails when the database cannot be read.
type Store interface {
	readings.Store
	verdict.Store
	explain.Store
	scenarios.Store
	ImportCount(ctx context.Context) (int, error)
}

// Models are the two models the explanation route asks, both nil when no provider key
// is configured.
type Models struct {
	Writer assist.Writer
	Judge  assist.Judge
}

// Config is what the server serves beside the storage and the models.
type Config struct {
	// Manifest names the sample readings, resolved against storage on each request.
	Manifest []scenarios.Entry
	// WebRoot is the built UI's directory.
	WebRoot string
	// Fingerprint identifies the model configuration the server runs with.
	Fingerprint string
}

// Routes builds the router: the API, the health check and the built UI.
func Routes(store Store, models Models, cfg Config) *gin.Engine {
	r := gin.New()
	// A redirect would answer an API path in HTML; unknown paths get the
	// not_found envelope instead.
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, rec any) {
		slog.Error("handler panicked", slog.Any("panic", rec), slog.String("path", c.Request.URL.Path))
		internalError(c)
	}))

	// The health check reads the database, so a server on a missing or locked file
	// reports unavailable with 503, and counts the imports, so a server on an empty
	// database says so. The fingerprint identifies the model configuration.
	r.GET("/_/health", func(c *gin.Context) {
		imports, err := store.ImportCount(c.Request.Context())
		if err != nil {
			slog.Error("health check could not read the database", slog.String("error", err.Error()))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "imports": nil, "fingerprint": cfg.Fingerprint, "error": "The database could not be read."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "imports": imports, "fingerprint": cfg.Fingerprint, "error": nil})
	})
	r.GET("/api/scenarios", func(c *gin.Context) {
		list, err := scenarios.Resolve(c.Request.Context(), store, cfg.Manifest)
		if respondIfFailed(c, err) {
			return
		}
		c.JSON(http.StatusOK, list)
	})
	r.GET("/api/readings/:id", func(c *gin.Context) {
		id, ok := readingIDParam(c)
		if !ok {
			return
		}
		detail, err := readings.Get(c.Request.Context(), store, id)
		if errors.Is(err, readings.ErrNotFound) {
			readingNotFound(c, id)
			return
		}
		if respondIfFailed(c, err) {
			return
		}
		c.JSON(http.StatusOK, detail)
	})
	r.POST("/api/reviews", func(c *gin.Context) {
		sub, ok := decodeSubmission(c)
		if !ok {
			return
		}
		out, err := verdict.Save(c.Request.Context(), store, sub)
		var invalid *verdict.InvalidError
		var miss *verdict.MissingError
		switch {
		case errors.As(err, &invalid):
			validationError(c, invalid.Messages...)
		case errors.As(err, &miss):
			readingNotFound(c, miss.ID)
		case respondIfFailed(c, err):
		default:
			c.JSON(http.StatusOK, out)
		}
	})

	assistRoutes(r, store, models)

	// The UI's files, and nothing else under the web root: a directory is never listed;
	// one holding an index.html is served, any other falls through to the 404 below.
	files := static.Serve("/", static.LocalFile(cfg.WebRoot, false))
	r.NoRoute(func(c *gin.Context) {
		if p := c.Request.URL.Path; p == "/api" || strings.HasPrefix(p, "/api/") {
			writeError(c, http.StatusNotFound, ErrorBody{Error: CodeNotFound, Message: "No API route at " + p + "."})
			return
		}
		// The file server answers /index.html with a redirect to /, so serve the page there.
		if c.Request.URL.Path == "/index.html" {
			c.Request.URL.Path = "/"
		}
		// The page names its scripts by a hash that changes with every build, while its
		// own modification time need not move, so a browser allowed to keep an old page
		// would load it and find no script. The page is always sent whole; the scripts,
		// hashed, may be kept.
		if c.Request.URL.Path == "/" {
			c.Request.Header.Del("If-Modified-Since")
			c.Request.Header.Del("If-None-Match")
			c.Header("Cache-Control", "no-store")
		}
		// Serving a file aborts the context; a request left unaborted found no file.
		if files(c); !c.IsAborted() {
			http.NotFound(c.Writer, c.Request)
		}
	})
	return r
}

func writeError(c *gin.Context, status int, body ErrorBody) {
	c.AbortWithStatusJSON(status, body)
}

// validationError answers 400 with one sentence per broken rule, joined into one
// message.
func validationError(c *gin.Context, sentences ...string) {
	writeError(c, http.StatusBadRequest, ErrorBody{Error: CodeValidation, Message: strings.Join(sentences, " ")})
}

// readingIDParam reads the reading ID in the path, answering 400 when it is not one.
func readingIDParam(c *gin.Context) (ids.ID, bool) {
	id, err := ids.Parse(c.Param("id"))
	if err != nil {
		validationError(c, "id "+ids.ReadingIDRule)
		return 0, false
	}
	return id, true
}

// readingNotFound names the reading that does not exist.
func readingNotFound(c *gin.Context, id ids.ID) {
	writeError(c, http.StatusNotFound, ErrorBody{Error: CodeReadingNotFound, Message: fmt.Sprintf("Reading %s does not exist.", id), MissingReadingIDs: []ids.ID{id}})
}

func internalError(c *gin.Context) {
	writeError(c, http.StatusInternalServerError, ErrorBody{Error: CodeInternal, Message: "The server could not complete the request."})
}

// respondIfFailed writes an err as a 500, if any, and reports whether it did. A
// request the client abandoned is logged and left unanswered.
func respondIfFailed(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		slog.Info("client abandoned the request", slog.String("path", c.Request.URL.Path))
		c.Abort()
		return true
	}
	slog.Error("request failed", slog.String("path", c.Request.URL.Path), slog.String("error", err.Error()))
	internalError(c)
	return true
}
