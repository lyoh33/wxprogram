package models

import "gorm.io/gorm"

type Course struct {
	ID             uint
	Name           string  `gorm:"varchar(255)"`
	Introduction   *string `gorm:"varchar(255);default:null"`
	PrefaceUrl     *string `gorm:"varchar(255);default:null"`
	EnrollmentCode *string `gorm:"varchar(64);default:null"`
	IsOpen         bool    `gorm:"bool; default:true"`
	UserID         *uint   `gorm:"default: null"`
}

func CreateCourse(db *gorm.DB, course Course) error {
	if err := db.AutoMigrate(&Course{}); err != nil {
		return err
	}
	db.Create(&course)
	return nil
}
