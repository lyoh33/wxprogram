package controllers

import (
	"fmt"
	"math/rand"
	"mio/gin-example/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 验证码缓存
var codeCache = make(map[string]codeInfo)

type codeInfo struct {
	code      string
	expiresAt time.Time
}

func HandleLogin(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求"})
		return
	}

	// 查找用户
	var user models.User
	if err := DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未注册"})
		return
	}

	// 检查审核状态
	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "账号未通过审核"})
		return
	}

	// 验证登录方式
	if req.Code != "" {
		if !verifyCode(req.Phone, req.Code) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码错误或已过期"})
			return
		}
	} else if req.Password != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择登录方式"})
		return
	}

	// 生成JWT
	token, err := generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	// 更新最后登录时间
	DB.Model(&user).Update("last_login", time.Now())

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"is_admin": user.Role,
		"user_id":  user.ID,
	})
}

// 处理发送验证码
func handleSendCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求"})
		return
	}

	code := generateRandomCode(6)
	expiration := time.Now().Add(5 * time.Minute)

	// 存储验证码
	codeCache[req.Phone] = codeInfo{
		code:      code,
		expiresAt: expiration,
	}

	// TODO: 集成短信发送服务
	fmt.Printf("发送验证码到 %s: %s\n", req.Phone, code)

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

// 生成随机验证码
func generateRandomCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}
	return string(code)
}

// 验证验证码
func verifyCode(phone, code string) bool {
	info, exists := codeCache[phone]
	if !exists {
		return false
	}

	if time.Now().After(info.expiresAt) {
		delete(codeCache, phone)
		return false
	}

	return info.code == code
}
