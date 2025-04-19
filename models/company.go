package models

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	Name        string  `gorm:"type:varchar(50);not null;uniqueIndex" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	Department  *string `gorm:"type:varchar(50)" json:"department,omitempty"`
}

func CreateCompany(db *gorm.DB, company Company) error {
	// err := db.Migrator().CreateTable(&Company{})
	err := db.AutoMigrate(&Company{})
	if err != nil {
		panic(err)
	}
	result := db.Create(company)
	return result.Error
}
