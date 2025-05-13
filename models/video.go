package models

import (
	"time"

	"gorm.io/gorm"
)

// 视频信息表
type Video struct {
	gorm.Model
	CourseID     uint      `gorm:"index"` // 外键
	UnitID       *uint     // 所属单元
	Title        string    `gorm:"type:varchar(100);not null"` // 视频标题
	URL          string    `gorm:"type:varchar(255);not null"` // 视频地址
	Duration     int       `gorm:"default:0"`                  // 时长（秒）
	IsMandatory  bool      `gorm:"default:true"`               // 是否必修
	WatchedCount uint      `gorm:"default:0"`                  // 观看次数
	LastWatched  time.Time // 最后观看时间
}
