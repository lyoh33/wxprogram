package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string `gorm:"type:varchar(50);not null"`
	Email        string `gorm:"type:varchar(255);unique;not null; index"`
	Phone        string `gorm:"type:varchar(20);unique;not null; index"`
	Age          *uint8 `gorm:"default:null"`
	Password     string `gorm:"type:varchar(255);not null"`
	TokenVersion uint   `gorm:"type:int;unsigned;default:0"`
	Role         string `gorm:"type:varchar(10);not null;default:'user';check:role IN ('user', 'admin')"`
	CompanyID    uint
	Company      Company  `gorm:"foreignKey:CompanyID"`
	Courses      []Course `gorm:"many2many:user_courses;"`
	Status       bool     `gorm:"default:true"`
}

func UserAutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func CreateUser(db *gorm.DB, userInput *User) (*User, error) {
	if err := db.Create(userInput).Error; err != nil {
		return nil, err
	}
	return userInput, nil
}

func Refresh(db *gorm.DB, id uint, version uint) {
	db.Model(&User{}).Where("id = ?", id).Update("token_version", version)
}

func FindUSerByEmail(db *gorm.DB, email string) *User {
	var user User
	db.Where("email = ?", email).First(&user)
	return &user
}

func FindUserByID(db *gorm.DB, id uint) *User {
	var user User
	db.Select("id", "name", "created_at").Where("id = ?", id).First(&user)
	return &user
}

func CreateUserIfNotExists(db *gorm.DB, userInput *User) (*User, bool, error) {
	var existingUser User

	// 优先通过 Email 判断用户是否存在（唯一性约束）
	result := db.Where("email = ?", userInput.Email).First(&existingUser)
	if result.Error != nil {
		// 用户不存在时创建
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(userInput).Error; err != nil {
				return nil, false, err // 创建失败（如唯一性冲突）
			}
			return userInput, true, nil // 创建成功
		}
		// 其他数据库错误
		return nil, false, result.Error
	}

	// 用户已存在
	return &existingUser, false, nil
}
