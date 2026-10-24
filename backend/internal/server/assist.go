package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"geolineai/internal/core/assist/explain"
	"geolineai/internal/core/review/readings"
)

// assistRoutes adds the explanation route. It writes nothing: model output is sent to
// the page and never stored, and every way a model can fail is an outcome of the
// answer rather than an error response.
func assistRoutes(r *gin.Engine, store Store, models Models) {
	r.POST("/api/readings/:id/explanations", func(c *gin.Context) {
		id, ok := readingIDParam(c)
		if !ok {
			return
		}
		body, ok := decodeStrings(c, "question")
		if !ok {
			return
		}
		out, err := explain.Ask(c.Request.Context(), store, models.Writer, models.Judge, id, body["question"])
		var invalid *explain.InvalidError
		switch {
		case errors.As(err, &invalid):
			validationError(c, invalid.Message)
		case errors.Is(err, readings.ErrNotFound):
			readingNotFound(c, id)
		case respondIfFailed(c, err):
		default:
			c.JSON(http.StatusOK, out)
		}
	})
}
