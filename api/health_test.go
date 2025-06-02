package api

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockScyllaSession is a mock implementation of gocql.Session
type MockScyllaSession struct {
	mock.Mock
}

func TestHealthCheckAPI(t *testing.T) {
	t.Skip("Skipping test: requires ScyllaDB setup")

	// Mock the CreateScyllaDBClient function to return a valid session
	CreateScyllaDBClient := func() (*gocql.Session, error) {
		return &gocql.Session{}, nil
	}

	_, _ = CreateScyllaDBClient()
	router := gin.Default()
	router.GET("/health", HealthCheckAPI)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponseBody := `{"message":"Employee API is not running. Check application logs"}`
	assert.Equal(t, expectedResponseBody, w.Body.String())
}

func TestHealthCheckAPI_Error(t *testing.T) {
	t.Skip("Skipping test: requires ScyllaDB setup")

	CreateScyllaDBClient := func() (*gocql.Session, error) {
		return nil, errors.New("ScyllaDB connection error")
	}
	_, _ = CreateScyllaDBClient()
	router := gin.Default()
	router.GET("/health", HealthCheckAPI)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponseBody := `{"message":"Employee API is not running. Check application logs"}`
	assert.Equal(t, expectedResponseBody, w.Body.String())
}

func TestGetRedisHealth(t *testing.T) {
	t.Skip("Skipping test: requires Redis setup")

	mockRedisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	mockRedisClient.Ping(ctx).Err()

	CreateRedisClient := func() *redis.Client {
		return mockRedisClient
	}
	CreateRedisClient()

	health := getRedisHealth()
	assert.Equal(t, "down", health)
}

func TestDetailedHealthCheckAPI(t *testing.T) {
	t.Skip("Skipping test: requires ScyllaDB and Redis setup")

	router := gin.Default()
	router.GET("/health", DetailedHealthCheckAPI)

	mockScyllaSession := &MockScyllaSession{}
	mockScyllaSession.On("Close").Return(nil)

	CreateScyllaDBClient := func() (*gocql.Session, error) {
		return &gocql.Session{}, nil
	}
	_, _ = CreateScyllaDBClient()

	getRedisHealth := func() string {
		return "up"
	}
	getRedisHealth()

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := performRequest(router, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedData := `{"message":"Employee API is not running. Check application logs","scylla_db":"down","employee_api":"down","redis":"down"}`
	assert.Equal(t, expectedData, w.Body.String())
}

func (m *MockScyllaSession) Close() error {
	args := m.Called()
	return args.Error(0)
}

func performRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
