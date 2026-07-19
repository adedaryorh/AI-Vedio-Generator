package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{Router: gin.New()}
	server.SetupRoutes()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	server.Router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
