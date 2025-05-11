package main

import (
	"errors"
	"fmt"
	"mio/gin-example/controllers"
	"mio/gin-example/middlewares"
	"mio/gin-example/models"
	"os"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
	// dsn := "mio@tcp(146.56.220.190:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "test.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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

	admin.GET("/company/:id", controllers.GetCompany)
	admin.GET("/company", controllers.GetCompanies)
	admin.POST("/company", controllers.CreateCompany)
	admin.PUT("/company/:id", controllers.UpdateCompany)
	admin.DELETE("company/:id", controllers.DeleteCompany)

	admin.GET("/user/:id", controllers.GetUser)
	admin.GET("/user", controllers.GetUsers)
	admin.PUT("/user/:id", controllers.UpdateUser)
	admin.DELETE("/user/:id", controllers.DeleteUser)

	admin.GET("/course/:id", controllers.GetCourse)
	admin.POST("/course", controllers.CreateCourse)
	admin.PUT("/course/:id", controllers.UpdateCourse)
	admin.DELETE("/course:id", controllers.DeleteCourse)

	admin.GET("/course/video", controllers.GetVideos)
	admin.GET("/course/video/:id", controllers.GetVideo)
	admin.POST("/course/video", controllers.CreateVideo)
	admin.PUT("/course/video/:id", controllers.UpdateVideo)

	admin.POST("/question_bank", controllers.CreateQuestionBank)
	admin.GET("/question_bank", controllers.GetQuestionBanks)


	r.POST("/v1/signup", controllers.Signup)
	r.POST("/v1/login", controllers.HandleLogin)
	r.GET("/course/:id", controllers.GetCourse)
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
