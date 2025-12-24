package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"comic-proxy/internal/storage"
)

type comicCommentRequest struct {
	Author  string `json:"author"`
	Content string `json:"content" binding:"required"`
}

func (h *Handler) registerComicRoutes(r *gin.Engine) {
	group := r.Group("/api/comics")
	group.POST("/:comicID/like", h.handleLikeComic)
	group.GET("/:comicID/comments", h.handleListComicComments)
	group.POST("/:comicID/comments", h.handleAddComicComment)
	group.DELETE("/comments/:commentID", h.handleDeleteComicComment)
	group.GET("/shared/featured", h.handleListFeaturedComics)
}

func (h *Handler) handleLikeComic(c *gin.Context) {
	comicIDStr := c.Param("comicID")
	comicID, err := strconv.ParseInt(comicIDStr, 10, 64)
	if err != nil || comicID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的漫画ID"})
		return
	}
	total, err := h.userStore.IncrementComicLikes(c.Request.Context(), comicID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrComicNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"likes": total})
}

func (h *Handler) handleAddComicComment(c *gin.Context) {
	comicIDStr := c.Param("comicID")
	comicID, err := strconv.ParseInt(comicIDStr, 10, 64)
	if err != nil || comicID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的漫画ID"})
		return
	}

	var req comicCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入留言内容"})
		return
	}

	comment, err := h.userStore.AddComicComment(
		c.Request.Context(),
		comicID,
		req.Author,
		req.Content,
	)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrComicNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

func (h *Handler) handleListComicComments(c *gin.Context) {
	comicIDStr := c.Param("comicID")
	comicID, err := strconv.ParseInt(comicIDStr, 10, 64)
	if err != nil || comicID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的漫画ID"})
		return
	}

	comments, err := h.userStore.ListComicComments(c.Request.Context(), comicID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrComicNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func (h *Handler) handleDeleteComicComment(c *gin.Context) {
	commentIDStr := c.Param("commentID")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil || commentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的留言ID"})
		return
	}

	if err := h.userStore.DeleteComicComment(c.Request.Context(), commentID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrCommentNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "留言已删除"})
}

func (h *Handler) handleListFeaturedComics(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	records, err := h.userStore.ListFeaturedComics(c.Request.Context(), limit, offset)
	if err != nil {
		status := http.StatusInternalServerError
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comics": records})
}
