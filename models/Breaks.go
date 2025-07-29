// internal/model/break_log.go
package models

import "time"

type BreakLog struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	AttendanceID uint       `gorm:"not null" json:"attendance_id"`
	BreakType    string     `gorm:"type:varchar(50);not null" json:"break_type"`
	BreakStart   time.Time  `gorm:"not null" json:"break_start"`
	BreakEnd     *time.Time `json:"break_end"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
