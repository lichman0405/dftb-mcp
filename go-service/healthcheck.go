package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("Health check passed")
		os.Exit(0)
	} else {
		fmt.Printf("Health check failed with status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}
}
