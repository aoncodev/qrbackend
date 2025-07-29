package models

import "time"

type AttendanceLog struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	EmployeeID uint       `gorm:"not null" json:"employee_id"`
	ClockIn    time.Time  `gorm:"not null" json:"clock_in"`
	ClockOut   *time.Time `json:"clock_out"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// ✅ Add this to link with breaks
	Breaks []BreakLog `gorm:"foreignKey:AttendanceID" json:"breaks"`
}
