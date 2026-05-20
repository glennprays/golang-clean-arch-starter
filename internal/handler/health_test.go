package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
	"github.com/glennprays/golang-clean-arch-starter/internal/testkit"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_ReturnsEnvelope(t *testing.T) {
	h := handler.NewHealthHandler()
	app := testkit.NewTestApp(t, func(app *fiber.App) {
		app.Get("/health", h.Check)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("X-Trace-Id"), "TraceID middleware should set X-Trace-Id")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var env struct {
		Data struct {
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
		} `json:"data"`
		TraceID string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(body, &env), "response body should be the standard envelope")

	assert.Equal(t, "ok", env.Data.Status)
	assert.NotEmpty(t, env.Data.Timestamp, "timestamp should be populated")
	assert.NotEmpty(t, env.TraceID, "trace_id should be populated in the envelope")
	assert.Equal(t, resp.Header.Get("X-Trace-Id"), env.TraceID,
		"trace_id in body must match X-Trace-Id header")
}
