package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string  `gorm:"type:varchar(50);not null"`
	Email        string  `gorm:"type:varchar(255);unique;not null; index"`
	Age          *uint8  `gorm:"default:null"`
	Password     string  `gorm:"type:varchar(255);not null"`
	TokenVersion uint    `gorm:"type:int;unsigned;default:0"`
	Role         *string `gorm:"type:ENUM('user', 'admin');not null;default:'user'"`
	CompanyID    int
	Company      Company
	Courses      []Course
	Status       uint //是否通过审核，0为不通过，1为通过
}

func UserAutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func CreateUser(db *gorm.DB, name, email, password, company string) {
	db.AutoMigrate(&User{})
	var _company Company
	db.Where("name = ?", company).First(&_company)
	user := User{
		Name:     name,
		Email:    email,
		Password: password,
		Company:  _company,
	}
	db.Create(&user)
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
