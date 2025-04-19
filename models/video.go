package models

import "gorm.io/gorm"

// 视频信息表
type Video struct {
	gorm.Model
	CourseID    uint    `gorm:"not null;index"`             // 所属课程ID
	UnitID      uint    `gorm:"not null;default:1"`         // 所属单元ID（默认第一单元）
	Title       string  `gorm:"type:varchar(255);not null"` // 视频标题
	URL         string  `gorm:"type:varchar(512);not null"` // 视频地址
	Duration    int     `gorm:"default:0"`                  // 视频时长（秒）
	Description *string `gorm:"type:text"`                  // 视频描述

	// 唯一约束：同一课程下视频标题不能重复
	Course Course     `gorm:"foreignKey:CourseID"`
	Unit   CourseUnit `gorm:"foreignKey:UnitID"`
}
