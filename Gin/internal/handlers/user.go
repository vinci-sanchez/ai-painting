package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"comic-proxy/internal/storage"
)

type userCredentialsRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) registerUserRoutes(r *gin.Engine) {
	group := r.Group("/api/users")
	group.POST("", h.handleRegisterUser)
	group.POST("/login", h.handleLoginUser)
	group.GET("/:username", h.handleGetUser)
}

func (h *Handler) handleRegisterUser(c *gin.Context) {
	var req userCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入有效的用户名和密码"})
		return
	}

	if err := h.userStore.RegisterUser(req.Username, req.Password); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, storage.ErrUserExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "用户已注册",
		"username": req.Username,
	})
}

func (h *Handler) handleLoginUser(c *gin.Context) {
	var req userCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入有效的用户名和密码"})
		return
	}

	if err := h.userStore.VerifyUser(req.Username, req.Password); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, storage.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "登录成功",
		"username": req.Username,
	})
}

func (h *Handler) handleGetUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	user, ok := h.userStore.GetUser(username)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到对应用户"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":   user.Username,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}
