package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "source"
	}

	fmt.Printf("[%s] Hello world - service starting up\n", serviceName)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Printf("[%s] heartbeat\n", serviceName)
	}
}
