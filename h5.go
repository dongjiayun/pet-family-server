package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	// Serve files from the "dist" directory
	router.NoRoute(func(c *gin.Context) {
		c.File("/etc/nginx/html/h5/index.html")
	})

	// Set "index.html" as the default file for the root path
	router.Static("/", "/etc/nginx/html/h5")

	router.Run(":8888")
}
