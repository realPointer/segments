package main

import (
	"github.com/realPointer/segments/internal/app"
)

// @title Dynamic user segmentation service
// @version 1.0.0
// @description A service that stores a user and the segments they are a part of (creating, modifying, deleting segments, and adding and deleting users to a segment)

// @host localhost:8080
// @BasePath /v1

// @contact.name Andrew
// @contact.url https://t.me/realPointer

func main() {
	// Run application
	app.Run()
}
