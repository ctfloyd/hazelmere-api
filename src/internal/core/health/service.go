package health

import (
	"context"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/version"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// DependencyStatus represents the health status of a single dependency
type DependencyStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "healthy", "unhealthy", "degraded"
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse is the complete health check response
type HealthResponse struct {
	Status       string             `json:"status"` // "healthy", "unhealthy", "degraded"
	Environment  string             `json:"environment"`
	Commit       string             `json:"commit"`
	BuildTime    string             `json:"buildTime"`
	Version      string             `json:"version"`
	Timestamp    string             `json:"timestamp"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

// Service performs health checks on application dependencies
type Service struct {
	mongoClient *mongo.Client
	dbName      string
	environment string
}

// NewService creates a new health service
func NewService(mongoClient *mongo.Client, dbName string, environment string) *Service {
	return &Service{
		mongoClient: mongoClient,
		dbName:      dbName,
		environment: environment,
	}
}

// Check performs a deep health check of all dependencies
// Returns the health response and whether the service is healthy (for HTTP status code)
func (s *Service) Check(ctx context.Context) (HealthResponse, bool) {
	info := version.Get()

	response := HealthResponse{
		Status:       "healthy",
		Environment:  s.environment,
		Commit:       info.Commit,
		BuildTime:    info.BuildTime,
		Version:      info.Version,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: make([]DependencyStatus, 0),
	}

	// Check MongoDB - this is critical
	mongoStatus := s.checkMongo(ctx)
	response.Dependencies = append(response.Dependencies, mongoStatus)

	// If MongoDB is unhealthy, the whole service is unhealthy
	isHealthy := true
	if mongoStatus.Status == "unhealthy" {
		response.Status = "unhealthy"
		isHealthy = false
	}

	return response, isHealthy
}

func (s *Service) checkMongo(ctx context.Context) DependencyStatus {
	status := DependencyStatus{
		Name:   "mongodb",
		Status: "healthy",
	}

	// Create a context with timeout for the ping
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	start := time.Now()
	err := s.mongoClient.Database(s.dbName).RunCommand(pingCtx, bson.D{{Key: "ping", Value: 1}}).Err()
	latency := time.Since(start)

	status.Latency = latency.Round(time.Millisecond).String()

	if err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
		return status
	}

	// Check if latency is too high (degraded)
	if latency > 1*time.Second {
		status.Status = "degraded"
	}

	return status
}
