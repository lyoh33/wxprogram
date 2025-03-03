package controllers

import (
	"mio/gin-example/models"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

func AddCompany(c *gin.Context) {
	name := c.PostForm("name")
	models.CreateCompany(DB, name)
}

func GetUser(c *gin.Context) {
	id := c.Params.ByName("id")
	_id, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}
	models.FindUserByID(DB, uint(_id))
}

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

func CreateCourse(c *gin.Context) {
	type CreateCourseRequest struct {
		Name           string `form:"name" binding:"required,max=255"`
		Introduction   string `form:"introduction" binding:"max=255"`
		PrefaceUrl     string `form:"preface_url" binding:"max=255"`
		EnrollmentCode string `form:"enrollment_code" binding:"max=64"`
		IsOpen         bool   `form:"is_open"`
	}
	var req CreateCourseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		log.Errorln(err)
		return
	}
	trimRequestFields(&req)
	courseModel := models.Course{
		Name: req.Name,
	}
	if req.Introduction != "" {
		courseModel.Introduction = &req.Introduction
	}
	if req.PrefaceUrl != "" {
		courseModel.PrefaceUrl = &req.PrefaceUrl
	}
	if req.EnrollmentCode != "" {
		courseModel.EnrollmentCode = &req.EnrollmentCode
	}
	courseModel.IsOpen = req.IsOpen

	err := models.CreateCourse(DB, courseModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}
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
