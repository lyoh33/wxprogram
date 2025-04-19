package controllers

import (
	"errors"
	"math"
	"mio/gin-example/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetUser(c *gin.Context) {
	// 参数校验中间件应已验证ID格式
	userID := c.Param("id")

	var user models.User
	result := DB.Model(&models.User{}).
		Select("id", "name", "email", "age", "role", "company_id", "created_at", "updated_at").
		Preload("Company", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Courses", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title", "progress")
		}).
		First(&user, userID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// 构造安全响应结构体
	type SafeUser struct {
		ID        uint      `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		Age       *uint8    `json:"age,omitempty"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	response := SafeUser{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Age:       user.Age,
		Role:      *user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GET /users?page=2&limit=10&sort=created_at desc&fields=name,email&role=admin
func GetUsers(c *gin.Context) {
	// 初始化查询
	query := DB.Model(&models.User{})

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "id")
	fields := c.Query("fields")

	// 参数验证
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit value (1-100)"})
		return
	}

	// 构建过滤条件
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if email := c.Query("email"); email != "" {
		query = query.Where("email = ?", email)
	}
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// 字段选择（白名单机制）
	validFields := map[string]bool{
		"id": true, "name": true, "email": true,
		"age": true, "role": true, "company_id": true,
		"created_at": true, "updated_at": true,
	}
	if fields != "" {
		selected := []string{"id"} // 保证始终包含ID
		for _, f := range strings.Split(fields, ",") {
			if validFields[f] {
				selected = append(selected, f)
			}
		}
		query = query.Select(selected)
	} else {
		// 默认排除敏感字段
		query = query.Omit("password", "token_version")
	}

	// 预加载关联数据
	query = query.Preload("Company", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name") // 只加载公司关键字段
	}).Preload("Courses")

	// 执行分页查询
	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	var users []models.User
	result := query.Order(sort).
		Limit(limit).
		Offset(offset).
		Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// 构造响应数据结构
	response := gin.H{
		"data": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	}

	c.JSON(http.StatusOK, response)

}

func UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	// 1. 获取现有用户数据
	var existingUser models.User
	if err := DB.First(&existingUser, userID).Error; err != nil {
		handleUserError(c, err)
		return
	}

	// 2. 解析请求体
	var updateData struct {
		Name     *string `json:"name" validate:"omitempty,min=2,max=50"`
		Email    *string `json:"email" validate:"omitempty,email"`
		Age      *uint8  `json:"age" validate:"omitempty,min=1,max=100"`
		Password *string `json:"password" validate:"omitempty,min=8"`
		Role     *string `json:"role" validate:"omitempty,oneof=user admin"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 3. 验证数据
	if err := validator.New().Struct(updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. 构建更新字段映射
	updates := make(map[string]any)
	if updateData.Name != nil {
		updates["name"] = *updateData.Name
	}
	if updateData.Email != nil {
		// 检查邮箱唯一性
		if exists := checkEmailExists(*updateData.Email, existingUser.ID); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		updates["email"] = *updateData.Email
	}
	if updateData.Age != nil {
		updates["age"] = updateData.Age
	}
	if updateData.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updateData.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
			return
		}
		updates["password"] = string(hashedPassword)
		updates["token_version"] = existingUser.TokenVersion + 1 // 令旧token失效
	}
	if updateData.Role != nil {
		// 添加角色修改权限检查（示例）
		// if !isAdmin(c) { ... }
		updates["role"] = *updateData.Role
	}

	// 5. 执行更新
	if err := DB.Model(&existingUser).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// 使用事务保证数据一致性
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. 检查用户存在性
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		// 2. 删除关联课程
		if err := tx.Model(&user).Association("Courses").Delete(); err != nil {
			return err
		}

		// 3. 执行软删除
		if err := tx.Delete(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete operation failed"})
		return
	}

	c.Status(http.StatusNoContent)
}

// 辅助函数
func handleUserError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
}

func checkEmailExists(email string, excludeID uint) bool {
	var count int64
	DB.Model(&models.User{}).
		Where("email = ? AND id != ?", email, excludeID).
		Count(&count)
	return count > 0
}
