package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	// Serve files from the "dist" directory
	router.NoRoute(func(c *gin.Context) {
		c.File("/etc/nginx/html/pc/index.html")
	})

	// Set "index.html" as the default file for the root path
	router.Static("/", "/etc/nginx/html/pc")

	router.Run(":9888")
}
