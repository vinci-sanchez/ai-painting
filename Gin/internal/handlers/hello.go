package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) registerHelloRoutes(r *gin.Engine) {
  r.GET("/api/hello", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "hello"})
  })
}