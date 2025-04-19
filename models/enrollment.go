package models

import (
	"time"

	"gorm.io/gorm"
)

// 课程报名信息主表
type Enrollment struct {
	gorm.Model
	UserID      uint       `gorm:"not null;index"`            // 用户ID
	CourseID    uint       `gorm:"not null;index"`            // 课程ID
	EnrolledAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP"` // 报名时间
	IsCompleted bool       `gorm:"default:false"`             // 是否完成课程
	CompletedAt *time.Time // 完成时间

	// 关联关系
	User          User                `gorm:"foreignKey:UserID"`
	Course        Course              `gorm:"foreignKey:CourseID"`
	VideoProgress []UserVideoProgress // 视频观看记录
	ExamRecords   []ExamRecord        // 考试记录
}

// 视频观看进度表
type UserVideoProgress struct {
	gorm.Model
	EnrollmentID uint      `gorm:"not null;index"` // 报名记录ID
	VideoID      uint      `gorm:"not null;index"` // 视频ID
	LastWatched  time.Time // 最后观看时间
	Progress     float64   `gorm:"default:0"`     // 观看进度百分比
	IsCompleted  bool      `gorm:"default:false"` // 是否看完

	// 关联关系
	Enrollment Enrollment `gorm:"foreignKey:EnrollmentID"`
	Video      Video      `gorm:"foreignKey:VideoID"`
}

// 考试记录表
type ExamRecord struct {
	gorm.Model
	EnrollmentID uint    `gorm:"not null;index"` // 报名记录ID
	Score        float64 `gorm:"default:0"`      // 考试成绩
	Passed       bool    `gorm:"default:false"`  // 是否通过
	// ExamData     JSON    `gorm:"type:json"`      // 考试详细数据

	Enrollment Enrollment `gorm:"foreignKey:EnrollmentID"`
}

