package models

import "gorm.io/gorm"

type Company struct {
	ID   int    `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(50);not null"`
}

func CreateCompany(db *gorm.DB, name string) {
	// err := db.Migrator().CreateTable(&Company{})
	err := db.AutoMigrate(&Company{})
	if err != nil {
		panic(err)
	}
	db.Create(&Company{Name: name})
}
