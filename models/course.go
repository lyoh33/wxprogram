package models

import (
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model
	Name            string       `gorm:"type:varchar(100);not null;uniqueIndex"` // 课程名称（唯一）
	Description     *string      `gorm:"type:text"`                              // 课程简介（可选）
	CoverImage      *string      `gorm:"type:varchar(255)"`                      // 课程封面URL（可选）
	Units           []CourseUnit // 课程单元列表
	EnrollmentCode  string       `gorm:"type:varchar(50);not null"`             // 报名验证码
	IsOpen          bool         `gorm:"default:false"`                         // 课程开放状态
	EnrollmentCount uint         `gorm:"default:0;check:enrollment_count >= 0"` // 报名人数统计
	CompletionCount uint         `gorm:"default:0;check:completion_count >= 0"` // 完成人数统计

	Exams []Exam
}

type CourseUnit struct {
	gorm.Model
	UnitName string `gorm:"type:varchar(100);not null"` // 单元名称
	Position int    `gorm:"default:0"`                  // 单元排序位置
}

func CreateCourse(db *gorm.DB, course Course) error {
	if err := db.AutoMigrate(&Course{}); err != nil {
		return err
	}
	db.Create(&course)
	return nil
}

// 增加报名人数
func IncrementEnrollment(db *gorm.DB, courseID uint) error {
	result := db.Model(&Course{}).
		Where("id = ?", courseID).
		UpdateColumn("enrollment_count",
			gorm.Expr("enrollment_count + ?", 1))

	return result.Error
}

// 增加完成人数（带前置校验）
func CompleteCourse(db *gorm.DB, courseID, userID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 检查用户是否已报名
		var enrollment Enrollment
		if err := tx.Where("user_id = ? AND course_id = ?",
			userID, courseID).First(&enrollment).Error; err != nil {
			return err
		}

		// 原子更新
		if err := tx.Model(&Course{}).
			Where("id = ?", courseID).
			UpdateColumn("completion_count",
				gorm.Expr("completion_count + 1")).Error; err != nil {
			return err
		}

		// 更新用户完成状态
		return tx.Model(&enrollment).Update("completed", true).Error
	})
}
