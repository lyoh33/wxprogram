package controllers

import (
	"mio/gin-example/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetVideos 获取视频列表
// @Summary 获取视频列表
// @Description 获取视频列表，支持分页和课程/单元过滤
// @Tags 视频管理
// @Produce json
// @Param course_id query int false "课程ID"
// @Param unit_id query int false "单元ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {array} models.Video
// @Router /admin/course/video [get]
func GetVideos(c *gin.Context) {
	var query struct {
		CourseID uint `form:"course_id"`
		UnitID   uint `form:"unit_id"`
		Page     int  `form:"page,default=1"`
		Limit    int  `form:"limit,default=10"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := DB.
		Preload("Course").
		Preload("Unit").
		Order("id desc")

	if query.CourseID > 0 {
		db = db.Where("course_id = ?", query.CourseID)
	}
	if query.UnitID > 0 {
		db = db.Where("unit_id = ?", query.UnitID)
	}

	var videos []models.Video
	offset := (query.Page - 1) * query.Limit
	if err := db.Offset(offset).Limit(query.Limit).Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		return
	}

	c.JSON(http.StatusOK, videos)
}

// GetVideo 获取单个视频
// @Summary 获取单个视频
// @Description 根据ID获取视频详情
// @Tags 视频管理
// @Produce json
// @Param id path int true "视频ID"
// @Success 200 {object} models.Video
// @Router /admin/course/video/{id} [get]
func GetVideo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var video models.Video
	if err := DB.
		Preload("Course").
		Preload("Unit").
		First(&video, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "视频不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		return
	}

	c.JSON(http.StatusOK, video)
}

// CreateVideoRequest 创建视频请求体
type CreateVideoRequest struct {
	CourseID    uint    `json:"course_id" binding:"required"`
	UnitID      uint    `json:"unit_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	URL         string  `json:"url" binding:"required"`
	Duration    int     `json:"duration"`
	Description *string `json:"description"`
}

// CreateVideo 创建视频
// @Summary 创建视频
// @Description 创建新视频
// @Tags 视频管理
// @Accept json
// @Produce json
// @Param body body CreateVideoRequest true "视频信息"
// @Success 201 {object} models.Video
// @Router /admin/course/video [post]
func CreateVideo(c *gin.Context) {
	var req CreateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	video := models.Video{
		CourseID:    req.CourseID,
		UnitID:      req.UnitID,
		Title:       req.Title,
		URL:         req.URL,
		Duration:    req.Duration,
		Description: req.Description,
	}

	if err := DB.Create(&video).Error; err != nil {
		if isDuplicateError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "同课程下视频标题不能重复"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建视频失败"})
		return
	}

	c.JSON(http.StatusCreated, video)
}

// UpdateVideoRequest 更新视频请求体
type UpdateVideoRequest struct {
	Title       *string `json:"title"`
	URL         *string `json:"url"`
	UnitID      *uint   `json:"unit_id"`
	Duration    *int    `json:"duration"`
	Description *string `json:"description"`
}

// UpdateVideo 更新视频
// @Summary 更新视频
// @Description 更新视频信息
// @Tags 视频管理
// @Accept json
// @Produce json
// @Param id path int true "视频ID"
// @Param body body UpdateVideoRequest true "更新信息"
// @Success 200 {object} models.Video
// @Router /admin/course/video/{id} [put]
func UpdateVideo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var video models.Video
	if err := DB.First(&video, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "视频不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		return
	}

	var req UpdateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData := make(map[string]interface{})
	if req.Title != nil {
		updateData["title"] = *req.Title
	}
	if req.URL != nil {
		updateData["url"] = *req.URL
	}
	if req.UnitID != nil {
		updateData["unit_id"] = *req.UnitID
	}
	if req.Duration != nil {
		updateData["duration"] = *req.Duration
	}
	if req.Description != nil {
		updateData["description"] = req.Description
	}

	if err := DB.Model(&video).Updates(updateData).Error; err != nil {
		if isDuplicateError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "同课程下视频标题不能重复"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新视频失败"})
		return
	}

	c.JSON(http.StatusOK, video)
}

// 辅助函数：判断是否为唯一约束错误
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// 根据数据库驱动类型判断错误类型
	// 这里简化处理，实际应根据具体数据库错误码判断
	return err.Error() == "UNIQUE constraint failed: videos.title"
}
