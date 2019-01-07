package utils

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

// CORSMiddleware sets cross-origin access to allow communication from any URL
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// CheckError is a function used to check for errors and log them if present.
func CheckError(err error, description string) {
	if err != nil {
		log.Fatalf(description, err)
		os.Exit(1)
	}
}
