package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// 考试主体表
type Exam struct {
	gorm.Model
	Name         string    `gorm:"type:varchar(100);not null;uniqueIndex"` // 考试名称
	Description  *string   `gorm:"type:text"`                              // 考试描述
	StartTime    time.Time `gorm:"not null"`                               // 考试开始时间
	EndTime      time.Time `gorm:"not null"`                               // 考试结束时间
	Duration     int       `gorm:"not null;check:duration > 0"`            // 考试时长（分钟）
	MaxAttempts  int       `gorm:"default:3;check:max_attempts > 0"`       // 最大尝试次数
	PassingScore int       `gorm:"default:60;check:passing_score >= 0"`    // 及格分数
	BelongsType  string    `gorm:"type:varchar(20);not null"`              // 所属类型 course/training
	BelongsID    uint      `gorm:"not null"`                               // 所属实体ID

	// 关联配置
	QuestionConfigs []ExamQuestionConfig `gorm:"foreignKey:ExamID"` // 试题配置
	Prerequisites   []ExamPrerequisite   `gorm:"foreignKey:ExamID"` // 前置课程
	ScoreRules      ExamScoreRule        `gorm:"foreignKey:ExamID"` // 分数规则
}

// 试题配置表
type ExamQuestionConfig struct {
	gorm.Model
	ExamID         uint   `gorm:"not null;index"`
	QuestionBankID uint   `gorm:"not null;index"`                   // 题库ID
	QuestionType   string `gorm:"type:varchar(20)"`                 // 题目类型（空表示不限）
	Amount         int    `gorm:"check:amount > 0"`                 // 抽题数量
	Difficulty     *int   `gorm:"check:difficulty BETWEEN 1 AND 5"` // 难度过滤

	Bank QuestionBank `gorm:"foreignKey:QuestionBankID"`
}

// 分数规则表
type ExamScoreRule struct {
	gorm.Model
	ExamID           uint    `gorm:"not null;uniqueIndex"`
	RuleType         string  `gorm:"type:varchar(20);not null"` // 规则类型：bank/type
	TargetID         uint    // 题库ID或题型标识
	ScorePerQuestion float64 `gorm:"not null;check:score_per_question >= 0"`
}

// 考试前置条件
type ExamPrerequisite struct {
	gorm.Model
	ExamID      uint `gorm:"not null;index"`
	CourseID    uint `gorm:"not null;index"`                                   // 必须完成的课程
	MinProgress int  `gorm:"default:100;check:min_progress BETWEEN 0 AND 100"` // 最小完成进度

	Course Course `gorm:"foreignKey:CourseID"`
}

// 自定义校验
func (e *Exam) Validate() error {
	if e.StartTime.After(e.EndTime) {
		return errors.New("考试开始时间不能晚于结束时间")
	}

	if len(e.QuestionConfigs) == 0 {
		return errors.New("必须配置试题")
	}

	if e.BelongsType != "course" && e.BelongsType != "training" {
		return errors.New("无效的所属类型")
	}

	return nil
}

// GORM钩子
func (e *Exam) BeforeCreate(tx *gorm.DB) error {
	return e.Validate()
}

type ExamAttempt struct {
	gorm.Model
	UserID    uint `gorm:"not null;index"`
	ExamID    uint `gorm:"not null;index"`
	StartTime time.Time
	EndTime   time.Time
	Score     float64
	IsPassed  bool

	// 关联关系
	User    User         `gorm:"foreignKey:UserID"`
	Exam    Exam         `gorm:"foreignKey:ExamID"`
	Answers []ExamAnswer `gorm:"foreignKey:AttemptID"`
}

type ExamAnswer struct {
	gorm.Model
	AttemptID  uint   `gorm:"not null;index"`
	QuestionID uint   `gorm:"not null;index"`
	UserAnswer string `gorm:"type:text"`
	IsCorrect  bool

	Question Question `gorm:"foreignKey:QuestionID"`
}
