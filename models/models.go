package models

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{}, &Company{}, &Course{}, &CourseUnit{}, &Video{})
}
