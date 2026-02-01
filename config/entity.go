package config

import (
	`time`

	`gorm.io/gorm`
)

const aegConfigTable = "aeg_config"

type ConfitEnt struct {
	ID        string         `json:"id" gorm:"column:id;primaryKey"`
	ConfigID  string         `json:"config_id" gorm:"column:config_id;uniqueIndex;not null"`
	RawValue  string         `json:"raw_value" gorm:"column:raw_value"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}
