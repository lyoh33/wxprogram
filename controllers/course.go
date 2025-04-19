package controllers

import (
	"mio/gin-example/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCourse(c *gin.Context) {
	type UnitInput struct {
		UnitName string `json:"unit_name" binding:"required"`
		Position int    `json:"position"`
	}

	type CourseInput struct {
		Name           string      `json:"name" binding:"required"`
		Description    *string     `json:"description"`
		CoverImage     *string     `json:"cover_image"`
		Units          []UnitInput `json:"units"`
		EnrollmentCode string      `json:"enrollment_code" binding:"required"`
		IsOpen         bool        `json:"is_open"`
	}

	var input CourseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查课程名称唯一性
	var existingCourse models.Course
	if result := DB.Where("name = ?", input.Name).First(&existingCourse); result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程名称已存在"})
		return
	}

	// 转换输入到模型
	course := models.Course{
		Name:           input.Name,
		Description:    input.Description,
		CoverImage:     input.CoverImage,
		EnrollmentCode: input.EnrollmentCode,
		IsOpen:         input.IsOpen,
	}

	// 转换单元
	var units []models.CourseUnit
	for _, u := range input.Units {
		units = append(units, models.CourseUnit{
			UnitName: u.UnitName,
			Position: u.Position,
		})
	}
	course.Units = units

	// 使用事务创建
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&course).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建课程失败"})
		return
	}

	c.JSON(http.StatusCreated, course)
}

func GetCourse(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程ID"})
		return
	}

	var course models.Course
	result := DB.Preload("Units", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).First(&course, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}

	c.JSON(http.StatusOK, course)
}

func UpdateCourse(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程ID"})
		return
	}

	type UpdateInput struct {
		Name           *string `json:"name"`
		Description    *string `json:"description"`
		CoverImage     *string `json:"cover_image"`
		EnrollmentCode *string `json:"enrollment_code"`
		IsOpen         *bool   `json:"is_open"`
	}

	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var course models.Course
	if result := DB.First(&course, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}

	// 更新字段
	if input.Name != nil {
		// 检查名称唯一性
		var existing models.Course
		if DB.Where("name = ? AND id != ?", *input.Name, id).First(&existing).RowsAffected > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "课程名称已存在"})
			return
		}
		course.Name = *input.Name
	}

	if input.Description != nil {
		course.Description = input.Description
	}

	if input.CoverImage != nil {
		course.CoverImage = input.CoverImage
	}

	if input.EnrollmentCode != nil {
		course.EnrollmentCode = *input.EnrollmentCode
	}

	if input.IsOpen != nil {
		course.IsOpen = *input.IsOpen
	}

	if err := DB.Save(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新课程失败"})
		return
	}

	c.JSON(http.StatusOK, course)
}

// todo：删除时需要删除移除视频、题库、参与学习的用户的信息等全部信息
func DeleteCourse(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程ID"})
		return
	}

	result := DB.Delete(&models.Course{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除课程失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "课程删除成功"})
}


