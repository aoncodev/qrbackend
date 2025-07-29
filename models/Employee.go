// internal/model/employee.go
package models

import "time"

type Employee struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"type:varchar(100);not null" json:"name"`
	QRID       string    `gorm:"type:varchar(50);unique;not null" json:"qr_id"`
	HourlyWage int       `gorm:"type:int;not null" json:"hourly_wage"`
	Role       string    `gorm:"type:varchar(20);not null" json:"role"`       // "admin" or "employee"
	StartTime  string    `gorm:"type:varchar(5);not null" json:"start_time"`  // stores time as "HH:MM"
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	OTP        string    `gorm:"column:otp" json:"otp"`
}

