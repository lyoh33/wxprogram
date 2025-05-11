package controllers

import (
	"mio/gin-example/models"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Signup(c *gin.Context) {
	type User struct {
		Name     string `form:"username" binding:"required,min=3,max=10"`
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required,password"`
		Company  string `form:"company" binding:"oneof=baidu google tencent"`
	}
	var user User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}
	u := models.FindUSerByEmail(DB, user.Email)
	if u.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email has been used",
		})
		log.Errorln("email has been used: ", user.Email)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}
	models.CreateUser(DB, user.Name, user.Email, string(hash), user.Company)
}

func Login(c *gin.Context) {
	type User struct {
		Email    string `form:"email" binding:"required,email" gorm:"unique"`
		Password string `form:"password" binding:"required,password"`
	}
	var user User

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}

	userInDB := models.FindUSerByEmail(DB, user.Email)
	if userInDB.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no this user",
		})
		log.Errorln("no this user: ", user.Email)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(userInDB.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}

	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userInDB.ID,
		// "pwd": userInDB.Password,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"rol": userInDB.Role,
		"ver": userInDB.TokenVersion + 1,
	}).SignedString([]byte("ygredgds"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}
	models.Refresh(DB, userInDB.ID, userInDB.TokenVersion+1)
	c.JSON(http.StatusOK, gin.H{
		"token": t,
	})
}

func trimRequestFields(req interface{}) {
	val := reflect.ValueOf(req).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// 处理普通字符串字段
		if field.Kind() == reflect.String {
			str := field.Interface().(string)
			field.SetString(strings.TrimSpace(str))
		}

		// 处理 *string 指针字段
		if field.Kind() == reflect.Ptr &&
			field.Type().Elem().Kind() == reflect.String &&
			!field.IsNil() {
			str := field.Elem().Interface().(string)
			trimmed := strings.TrimSpace(str)
			field.Elem().SetString(trimmed)
		}
	}
}

func generateJWT(user models.User) (any, any) {
	userInDB := models.FindUSerByEmail(DB, user.Email)
	if userInDB.ID == 0 {
		return nil, "no this user"
	}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userInDB.ID,
		// "pwd": userInDB.Password,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"rol": userInDB.Role,
		"ver": userInDB.TokenVersion + 1,
	}).SignedString([]byte("ygredgds"))
	if err != nil {
		return nil, err
	}
	models.Refresh(DB, userInDB.ID, userInDB.TokenVersion+1)
	return t, nil
}
