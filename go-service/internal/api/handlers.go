package api

import (
	"fmt"
	"net/http"
	"time"
	"dftbopt-mcp/go-service/internal/dftb"
	"dftbopt-mcp/go-service/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIHandler handles HTTP API requests
type APIHandler struct {
	dftbRunner *dftb.DFTBRunner
	config     *types.ServerConfig
}

// NewAPIHandler creates a new API handler instance
func NewAPIHandler(config *types.ServerConfig) *APIHandler {
	return &APIHandler{
		dftbRunner: dftb.NewDFTBRunner(config),
		config:     config,
	}
}

// RegisterRoutes registers all API routes
func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.healthCheck)
	
	// Service info
	router.GET("/info", h.getServiceInfo)
	
	// Optimization endpoint
	router.POST("/api/v1/optimize", h.optimizeStructure)
	
	// Status check endpoint
	router.GET("/api/v1/status/:requestID", h.getStatus)
	
	// Middleware
	router.Use(h.corsMiddleware())
	router.Use(h.requestIDMiddleware())
	router.Use(h.loggingMiddleware())
}

// healthCheck handles health check requests
func (h *APIHandler) healthCheck(c *gin.Context) {
	response := types.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

// getServiceInfo returns service information
func (h *APIHandler) getServiceInfo(c *gin.Context) {
	response := types.ServiceInfo{
		Name:    "DFTB+ Optimization Service",
		Version: "1.0.0",
		Capabilities: []string{
			"geometry_optimization",
			"gfn1_xtb",
			"gfn2_xtb",
			"cif_input",
			"cif_output",
		},
	}
	c.JSON(http.StatusOK, response)
}

// optimizeStructure handles optimization requests
func (h *APIHandler) optimizeStructure(c *gin.Context) {
	var request types.OptimizationRequest
	
	// Bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}
	
	// Generate request ID if not provided
	if request.RequestID == "" {
		request.RequestID = uuid.New().String()
	}
	
	// Validate request
	if err := h.dftbRunner.ValidateRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}
	
	// Run optimization (in a real implementation, this should be async)
	response, err := h.dftbRunner.RunOptimization(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Optimization failed",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// getStatus returns the status of a calculation
func (h *APIHandler) getStatus(c *gin.Context) {
	requestID := c.Param("requestID")
	
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request ID is required",
		})
		return
	}
	
	status, err := h.dftbRunner.GetStatus(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get status",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"request_id": requestID,
		"status":     status,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// corsMiddleware handles CORS headers
func (h *APIHandler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		
		c.Next()
	}
}

// requestIDMiddleware adds request ID to context
func (h *APIHandler) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		
		c.Next()
	}
}

// loggingMiddleware logs request information
func (h *APIHandler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		requestID, _ := c.Get("requestID")
		
		gin.DefaultWriter.Write([]byte(fmt.Sprintf(
			"[%s] %s %s %s %d %v\n",
			time.Now().Format(time.RFC3339),
			c.Request.Method,
			c.Request.URL.Path,
			requestID,
			c.Writer.Status(),
			duration,
		)))
	}
}

// errorResponse creates a standardized error response
func (h *APIHandler) errorResponse(c *gin.Context, statusCode int, message string, details interface{}) {
	c.JSON(statusCode, gin.H{
		"error":   message,
		"details": details,
		"request_id": c.GetString("requestID"),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// successResponse creates a standardized success response
func (h *APIHandler) successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"request_id": c.GetString("requestID"),
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}
