package models

import (
	"errors"

	"gorm.io/gorm"
)

type Course struct {
	gorm.Model
	Name            string       `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description     string       `gorm:"type:text;not null"`
	CoverImage      string       `gorm:"type:varchar(255)"`
	Units           []CourseUnit `gorm:"foreignKey:CourseID"`
	EnrollmentCode  string       `gorm:"type:varchar(50);not null;uniqueIndex"`
	IsOpen          bool         `gorm:"default:false"`
	EnrollmentCount uint         `gorm:"default:0"`
	CompletionCount uint         `gorm:"default:0"`
	Videos          []Video      `gorm:"foreignKey:CourseID"`
	// 移除 ExamID，改为在 Exam 中关联 Course
}

type CourseUnit struct {
	gorm.Model
	UnitName    string `gorm:"type:varchar(100);not null"`
	Description string `gorm:"type:text"`
	Order       int    `gorm:"default:0;index:idx_course_order,unique"`
	CourseID    uint   `gorm:"index:idx_course_order,unique"`
}

type CourseService struct {
	db *gorm.DB
}

func NewCourseService(db *gorm.DB) *CourseService {
	return &CourseService{db: db}
}

// CreateCourse 创建课程
func (s *CourseService) CreateCourse(course *Course) error {
	// 检查课程名称唯一性
	var count int64
	s.db.Model(&Course{}).Where("name = ?", course.Name).Count(&count)
	if count > 0 {
		return errors.New("课程名称已存在")
	}

	// 自动生成纯色封面（示例逻辑）
	if course.CoverImage == "" {
		defaultCover := generateDefaultCover(course.Name)
		course.CoverImage = defaultCover
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(course).Error; err != nil {
			return err
		}

		// 创建默认第一单元
		if len(course.Units) == 0 {
			defaultUnit := CourseUnit{
				UnitName:    "课程导论",
				Description: "课程介绍和基本要求",
				Order:       1,
			}
			if err := tx.Create(&defaultUnit).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// IncrementEnrollment 增加报名人数（原子操作）
func (s *CourseService) IncrementEnrollment(courseID uint) error {
	return s.db.Model(&Course{}).
		Where("id = ?", courseID).
		UpdateColumn("enrollment_count", gorm.Expr("enrollment_count + 1")).
		Error
}

// IncrementCompletion 增加完成人数（原子操作）
func (s *CourseService) IncrementCompletion(courseID uint) error {
	return s.db.Model(&Course{}).
		Where("id = ?", courseID).
		UpdateColumn("completion_count", gorm.Expr("completion_count + 1")).
		Error
}

// 生成默认封面（示例实现）
func generateDefaultCover(name string) string {
	// 实际应实现生成纯色背景+文字的逻辑
	return ""
}
