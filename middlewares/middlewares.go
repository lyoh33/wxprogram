package middlewares

import (
	"fmt"
	"mio/gin-example/controllers"
	"mio/gin-example/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

func AdminRequired(c *gin.Context) {
	token := c.Request.Header.Get("token")
	claims, err := parseToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}
	ok, err := CheckVersion(claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		log.Errorln(err)
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid login token",
		})
		log.Errorln("bad token: ", token)
		return
	}
	if IsExpired(claims) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token is outdate",
		})
		log.Errorln("bad token", token)
		return
	}
	if claims.Rol != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "permission denied",
		})
		log.Errorln("permission denied", claims.Rol)
		return
	}
	c.Next()
}

func CheckVersion(claims *CustomClaims) (bool, error) {
	user := models.FindUserByID(controllers.DB, claims.Sub)
	_ver, err := strconv.Atoi(claims.Ver)
	if err != nil {
		return false, err
	}
	ver := uint(_ver)
	return ver <= user.TokenVersion, nil
}

func IsExpired(claims *CustomClaims) bool {
	// user := models.FindUserByID(controllers.DB, claims.Sub)
	exp := time.Unix(claims.Exp, 0)
	return time.Now().After(exp)
}

type CustomClaims struct {
	Sub uint   `json:"sub"`
	Exp int64  `json:"exp"`
	Rol string `json:"rol"`
	Ver string `json:"ver"`
	jwt.RegisteredClaims
}

func parseToken(tokenString string) (*CustomClaims, error) {
	// 解析Token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 返回密钥
		return []byte("ygredgds"), nil
	})

	// 处理解析错误
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// 验证Claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func Logger(c *gin.Context) {
	start := time.Now()
	c.Next()
	end := time.Now()
	log.WithFields(log.Fields{
		"method":  c.Request.Method,
		"path":    c.Request.URL.Path,
		"status":  c.Writer.Status(),
		"latency": end.Sub(start),
	}).Info("client request: ")
}
