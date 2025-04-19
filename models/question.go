package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// 自定义 JSON 类型（通用处理方案）
type JSONB map[string]any

func (j *JSONB) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言失败")
	}
	return json.Unmarshal(bytes, &j)
}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// 题库表（保持不变）
type QuestionBank struct {
	gorm.Model
	Name         string `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description  string `gorm:"type:text"`
	QuestionType string `gorm:"type:varchar(20);not null;check:question_type IN ('single_choice', 'multiple_choice', 'fill_blank', 'subjective')"`
}

// 题目表（修正版）
type Question struct {
	gorm.Model
	BankID     uint   `gorm:"not null;index"`
	Type       string `gorm:"type:varchar(20);not null"`
	Content    string `gorm:"type:text;not null"`
	Difficulty int    `gorm:"default:3;check:difficulty BETWEEN 1 AND 5"`
	Score      int    `gorm:"default:10"`

	// 使用自定义 JSON 类型
	Options    JSONB  `gorm:"type:json"`          // 选择题选项
	Answers    JSONB  `gorm:"type:json;not null"` // 正确答案
	Candidates JSONB  `gorm:"type:json"`          // 填空题候选答案
	Reference  string `gorm:"type:text"`          // 主观题参考答案
	Analysis   string `gorm:"type:text"`          // 题目解析

	Bank QuestionBank `gorm:"foreignKey:BankID"`
}
