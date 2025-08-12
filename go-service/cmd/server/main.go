package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"dftbopt-mcp/go-service/internal/api"
	"dftbopt-mcp/go-service/internal/types"
	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	var (
		port        = flag.Int("port", 8080, "Server port")
		workDir     = flag.String("work-dir", "./work", "Working directory for calculations")
		dftbPath    = flag.String("dftb-path", "dftb+", "Path to DFTB+ executable")
		maxRequests = flag.Int("max-requests", 10, "Maximum concurrent requests")
		timeout     = flag.Int("timeout", 300, "Calculation timeout in seconds")
		debug       = flag.Bool("debug", false, "Enable debug mode")
		cleanup     = flag.Bool("cleanup", false, "Enable automatic cleanup of old files")
	)
	flag.Parse()

	// Set Gin mode
	if *debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create server configuration
	config := &types.ServerConfig{
		Port:        *port,
		WorkDir:     *workDir,
		DFTBPath:    *dftbPath,
		MaxRequests: *maxRequests,
		Timeout:     *timeout,
	}

	// Create working directory if it doesn't exist
	if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
		log.Fatalf("Failed to create working directory: %v", err)
	}

	// Create API handler
	apiHandler := api.NewAPIHandler(config)

	// Create Gin router
	router := gin.New()
	
	// Register routes
	apiHandler.RegisterRoutes(router)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting DFTB+ Optimization Server on port %d", config.Port)
		log.Printf("Working directory: %s", config.WorkDir)
		log.Printf("DFTB+ executable: %s", config.DFTBPath)
		log.Printf("Max concurrent requests: %d", config.MaxRequests)
		log.Printf("Calculation timeout: %d seconds", config.Timeout)
		log.Printf("Debug mode: %v", *debug)
		log.Printf("Automatic cleanup: %v", *cleanup)

		if err := router.Run(fmt.Sprintf(":%d", config.Port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start cleanup routine if enabled
	if *cleanup {
		go func() {
			cleanupTicker := time.NewTicker(1 * time.Hour)
			defer cleanupTicker.Stop()

			for {
				select {
				case <-cleanupTicker.C:
					if err := cleanupOldFiles(config.WorkDir, 24*time.Hour); err != nil {
						log.Printf("Cleanup failed: %v", err)
					} else {
						log.Printf("Cleanup completed successfully")
					}
				}
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Here you would add any cleanup logic for your application
	log.Println("Server exited")
}

// cleanupOldFiles removes files older than the specified age
func cleanupOldFiles(workDir string, maxAge time.Duration) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return fmt.Errorf("failed to read working directory: %v", err)
	}

	now := time.Now()
	cleanedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(workDir, entry.Name())
			
			// Get directory info
			info, err := entry.Info()
			if err != nil {
				log.Printf("Failed to get info for directory %s: %v", entry.Name(), err)
				continue
			}

			// Check if directory is older than maxAge
			if now.Sub(info.ModTime()) > maxAge {
				if err := os.RemoveAll(dirPath); err != nil {
					log.Printf("Failed to remove directory %s: %v", entry.Name(), err)
				} else {
					cleanedCount++
					log.Printf("Cleaned up old directory: %s", entry.Name())
				}
			}
		}
	}

	log.Printf("Cleanup completed: removed %d old directories", cleanedCount)
	return nil
}

// printBanner prints the server banner
func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║                    DFTB+ OPTIMIZATION SERVICE                               ║
║                                                                              ║
║  A high-performance service for DFTB+ geometry optimization calculations       ║
║  supporting GFN1-xTB and GFN2-xTB methods with CIF file I/O.                  ║
║                                                                              ║
║  Features:                                                                   ║
║  • RESTful API for optimization requests                                      ║
║  • Automatic file management and cleanup                                      ║
║  • Concurrent request handling                                                ║
║  • Comprehensive error handling and logging                                  ║
║                                                                              ║
║  API Endpoints:                                                              ║
║  • GET  /health        - Health check                                        ║
║  • GET  /info          - Service information                                 ║
║  • POST /api/v1/optimize - Submit optimization request                        ║
║  • GET  /api/v1/status/:id - Check calculation status                        ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// validateConfig validates the server configuration
func validateConfig(config *types.ServerConfig) error {
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", config.Port)
	}

	if config.WorkDir == "" {
		return fmt.Errorf("working directory cannot be empty")
	}

	if config.DFTBPath == "" {
		return fmt.Errorf("DFTB+ executable path cannot be empty")
	}

	if config.MaxRequests <= 0 {
		return fmt.Errorf("max requests must be positive")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// Check if DFTB+ executable exists
	if _, err := os.Stat(config.DFTBPath); os.IsNotExist(err) {
		return fmt.Errorf("DFTB+ executable not found at: %s", config.DFTBPath)
	}

	return nil
}

// setupLogging configures logging
func setupLogging(debug bool) {
	if debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}
}

// createPIDFile creates a PID file for the server
func createPIDFile(pidFile string) error {
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

// removePIDFile removes the PID file
func removePIDFile(pidFile string) {
	os.Remove(pidFile)
}
