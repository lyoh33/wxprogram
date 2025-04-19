package controllers

import (
	"errors"
	"mio/gin-example/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCompany(c *gin.Context) {
	// 1. 声明接收请求的结构体
	var company models.Company

	// 2. 绑定 JSON 请求体
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// 3. 验证必要字段（应用层验证）
	if company.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "公司名称不能为空"})
		return
	}

	// 4. 创建数据库记录
	err := models.CreateCompany(DB, company)
	if err != nil {
		// 处理具体错误类型
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "公司ID已存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		}
		return
	}

	// 5. 返回创建成功的响应（HTTP 201）
	c.JSON(http.StatusCreated, gin.H{
		"message": "公司创建成功",
		"data":    company,
	})
}

func GetCompany(c *gin.Context) {
	// 从URL中提取ID
	id := c.Param("id")

	var company models.Company
	// 使用GORM查询（包含软删除记录）
	result := DB.First(&company, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "company not found",
			})
			return
		}
		// 其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error,
		})
		return
	}

	// 成功响应（自动处理omitempty）
	c.JSON(http.StatusOK, company)
}

func GetCompanies(c *gin.Context) {
	var companies []*models.Company
	// 使用GORM查询（包含软删除记录）
	result := DB.Find(&companies)

	if result.Error != nil {
		// 其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error,
		})
		return
	}

	// 成功响应（自动处理omitempty）
	c.JSON(http.StatusOK, companies)
}

func UpdateCompany(c *gin.Context) {
	// 从URL中提取ID
	id := c.Param("id")

	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// 使用GORM查询（包含软删除记录）
	result := DB.Model(&company).Where("id = ?", id).Updates(&company)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "company not found",
			})
			return
		}
		// 其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error,
		})
		return
	}

	// 成功响应（自动处理omitempty）
	c.JSON(http.StatusOK, company)
}

func DeleteCompany(c *gin.Context) {
	// 从URL中提取ID
	id := c.Param("id")

	// 使用GORM查询（包含软删除记录）
	result := DB.Delete(&models.Company{}, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "company not found",
			})
			return
		}
		// 其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}
