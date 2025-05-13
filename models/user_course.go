package models

import "time"

type UserCourse struct {
	UserID      uint       `gorm:"primaryKey;autoIncrement:false"` // 用户ID（联合主键）
	CourseID    uint       `gorm:"primaryKey;autoIncrement:false"` // 课程ID（联合主键）
	EnrolledAt  time.Time  `gorm:"autoCreateTime"`                 // 报名时间（自动记录）
	IsCompleted bool       `gorm:"default:false"`                  // 是否完成课程
	CompletedAt *time.Time // 完成时间（可空）
	Progress    uint       `gorm:"default:0"` // 学习进度百分比（0-100）
}
