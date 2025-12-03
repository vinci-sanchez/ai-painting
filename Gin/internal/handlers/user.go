package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"comic-proxy/internal/storage"
)

type userCredentialsRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type userComicRequest struct {
	Title       string                 `json:"title" binding:"required"`
	PageNumber  int                    `json:"page_number"`
	ImageBase64 string                 `json:"image_base64"`
	ImageURL    string                 `json:"image_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func (h *Handler) registerUserRoutes(r *gin.Engine) {
	group := r.Group("/api/users")
	group.POST("", h.handleRegisterUser)
	group.POST("/login", h.handleLoginUser)
	group.GET("/:username/comics", h.handleListUserComics)
	group.POST("/:username/comics", h.handleCreateUserComic)
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
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func (h *Handler) handleCreateUserComic(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var req userComicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入完整的漫画信息"})
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题不能为空"})
		return
	}

	imageBase64 := strings.TrimSpace(req.ImageBase64)
	if imageBase64 == "" {
		imageURL := strings.TrimSpace(req.ImageURL)
		if imageURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少漫画图片"})
			return
		}
		converted, err := h.fetchImageAsBase64(c.Request.Context(), imageURL)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("无法下载图片: %v", err)})
			return
		}
		imageBase64 = converted
	}

	var metadata json.RawMessage
	if len(req.Metadata) > 0 {
		raw, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "metadata 字段不是有效的 JSON"})
			return
		}
		metadata = raw
	}

	record, err := h.userStore.SaveComicForUser(
		c.Request.Context(),
		username,
		storage.ComicRecord{
			Title:       req.Title,
			PageNumber:  req.PageNumber,
			ImageBase64: imageBase64,
			Metadata:    metadata,
		},
	)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrUserNotFound) || errors.Is(err, storage.ErrEmptyUsername) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"comic": record,
	})
}

func (h *Handler) handleListUserComics(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	comics, err := h.userStore.ListComicsForUser(c.Request.Context(), username)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comics": comics})
}

func (h *Handler) fetchImageAsBase64(ctx context.Context, imageURL string) (string, error) {
	url := strings.TrimSpace(imageURL)
	if url == "" {
		return "", errors.New("image url is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}
	encoded := base64.StdEncoding.EncodeToString(body)
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded), nil
}
