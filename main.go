package main

import (
	"errors"
	"fmt"
	"mio/gin-example/controllers"
	"mio/gin-example/middlewares"
	"mio/gin-example/models"
	"net/http"
	"os"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Article struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var articles = []Article{
	{Id: "0", Title: "ab", Content: "hoiadoawdadl;am"},
	{Id: "1", Title: "dwq", Content: "sad213ewq"},
	{Id: "2", Title: "xc", Content: "zxchtrewsfas"},
}

func GetArticles(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, articles)
}

func PostArticle(c *gin.Context) {
	var article Article
	if err := c.BindJSON(&article); err != nil {
		fmt.Println(err)
		return
	}
	articles = append(articles, article)
	c.IndentedJSON(http.StatusOK, articles)
}

func GetArticle(c *gin.Context) {
	id := c.Param("id")

	for _, article := range articles {
		if article.Id == id {
			c.IndentedJSON(http.StatusOK, article)
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "article not found"})
}

func verifyPassword(s string) (fiveOrMore, number, upper, special bool, err error) {
	letters := 0
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
			letters++
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c):
			if c == '.' {
				special = true
				letters++
			} else {
				err = errors.New("unexpected punctation")
			}
		case unicode.IsLetter(c):
			letters++
		default:
			//return false, false, false, false
		}
	}
	fiveOrMore = letters >= 5
	return
}

func passwordValidate(fl validator.FieldLevel) bool {
	if password, ok := fl.Field().Interface().(string); !ok {
		panic("error")
	} else {
		f, n, u, s, err := verifyPassword(password)
		if err != nil {
			panic(err)
		}
		if f && n && u && s {
			return true
		}
	}
	return false
}

func InitLogrus() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	dsn := "mio@tcp(146.56.220.190:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	controllers.DB = db

	// r := gin.Default()
	InitLogrus()
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.Logger)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", passwordValidate)
	}
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	models.AutoMigrate(controllers.DB)

	admin := r.Group("/v1/admin")
	admin.Use(middlewares.AdminRequired)
	admin.GET("/user/:id", controllers.GetUser)
	admin.POST("/course", controllers.CreateCourse)

	r.POST("/v1/company", controllers.AddCompany)
	r.POST("/v1/signup", controllers.Signup)
	r.POST("/v1/login", controllers.Login)
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
