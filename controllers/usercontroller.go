package controllers

import (
	"net/http"
	"time"

	"github.com/aoncodev/qrbackend/initializers"
	"github.com/aoncodev/qrbackend/models"
	"github.com/gin-gonic/gin"
)

type EmployeeStatusRequest struct {
	QRID string `json:"qr_id" binding:"required"`
}

type CurrentBreak struct {
	ID         uint   `json:"id"`
	BreakType  string `json:"break_type"`
	BreakStart string `json:"break_start"` // ISO8601 string
}

type EmployeeStatusResponse struct {
	EmployeeID           uint         `json:"employee_id"`
	EmployeeName         string       `json:"employee_name"`
	Status               string       `json:"status"` // "not_clocked_in", "working", "on_break", "clocked_out"
	CurrentAttendanceID  *uint        `json:"current_attendance_id,omitempty"`
	ClockInTime          *string      `json:"clock_in_time,omitempty"`
	CurrentBreak         *CurrentBreak `json:"current_break,omitempty"`
}



func GetEmployeeStatus(c *gin.Context) {
	var req EmployeeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR ID is required"})
		return
	}

	var employee models.Employee
	if err := initializers.DB.Where("qr_id = ?", req.QRID).First(&employee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	var attendance models.AttendanceLog
	err := initializers.DB.
		Where("employee_id = ? AND clock_out IS NULL", employee.ID).
		Order("created_at DESC").
		First(&attendance).Error

	if err != nil {
		// No attendance found — not clocked in yet
		c.JSON(http.StatusOK, EmployeeStatusResponse{
			EmployeeID:   employee.ID,
			EmployeeName: employee.Name,
			Status:       "not_clocked_in",
		})
		return
	}

	var breakLog models.BreakLog
	breakErr := initializers.DB.
		Where("attendance_id = ? AND break_end IS NULL", attendance.ID).
		Order("created_at DESC").
		First(&breakLog).Error

	var status string = "working"
	var currentBreak * CurrentBreak = nil
	if breakErr == nil {
		status = "on_break"
		currentBreak = &CurrentBreak{
			ID:         breakLog.ID,
			BreakType:  breakLog.BreakType,
			BreakStart: breakLog.BreakStart.Format(time.RFC3339),
		}
	}

	clockInStr := attendance.ClockIn.Format(time.RFC3339)
	c.JSON(http.StatusOK, EmployeeStatusResponse{
		EmployeeID:          employee.ID,
		EmployeeName:        employee.Name,
		Status:              status,
		CurrentAttendanceID: &attendance.ID,
		ClockInTime:         &clockInStr,
		CurrentBreak:        currentBreak,
	})
}


func EmployeeLogin(c *gin.Context) {
	var req struct {
		QRID string `json:"qr_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR ID required"})
		return
	}

	var employee models.Employee
	if err := initializers.DB.Where("qr_id = ?", req.QRID).First(&employee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   employee.ID,
		"name": employee.Name,
		"role": employee.Role,
	})
}


func GetEmployeeStatusByID(c *gin.Context) {
	employeeID := c.Param("id")

	var employee models.Employee
	if err := initializers.DB.First(&employee, employeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Load today’s attendance (whether or not it's clocked out)
	var attendance models.AttendanceLog
	err := initializers.DB.
		Where("employee_id = ? AND DATE(clock_in) = ?", employee.ID, time.Now().UTC().Format("2006-01-02")).
		Preload("Breaks").
		First(&attendance).Error

	if err != nil {
		// No attendance today at all
		c.JSON(http.StatusOK, gin.H{
			"attendance": nil,
			"breaks":     nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"attendance": attendance,
		"breaks":     attendance.Breaks,
	})
}

func ClockIn(c *gin.Context) {
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required"})
		return
	}

	var employee models.Employee
	if err := initializers.DB.First(&employee, employeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Enforce one shift per day: check if clock-in already exists today
	var existing models.AttendanceLog
	err := initializers.DB.
		Where("employee_id = ? AND DATE(clock_in) = ?", employee.ID, time.Now().UTC().Format("2006-01-02")).
		First(&existing).Error

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already clocked in today"})
		return
	}

	// Create new attendance log
	newAttendance := models.AttendanceLog{
		EmployeeID: employee.ID,
		ClockIn:    time.Now(),
	}

	if err := initializers.DB.Create(&newAttendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clock in"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Clock-in successful",
		"attendance_id": newAttendance.ID,
		"clock_in":      newAttendance.ClockIn,
	})
}



func ClockOut(c *gin.Context) {
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required"})
		return
	}

	var employee models.Employee
	if err := initializers.DB.First(&employee, employeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Find open attendance log
	var attendance models.AttendanceLog
	err := initializers.DB.
		Where("employee_id = ? AND clock_out IS NULL", employee.ID).
		Order("created_at DESC").
		First(&attendance).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active attendance log found"})
		return
	}

	// Check if currently on a break
	var activeBreak models.BreakLog
	err = initializers.DB.
		Where("attendance_id = ? AND break_end IS NULL", attendance.ID).
		First(&activeBreak).Error

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You must end your break before clocking out"})
		return
	}

	// Clock out
	now := time.Now()
	attendance.ClockOut = &now

	if err := initializers.DB.Save(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clock out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Clock-out successful",
		"attendance_id": attendance.ID,
		"clock_out":     attendance.ClockOut,
	})
}


func StartBreak(c *gin.Context) {
	var req struct {
		AttendanceID uint   `json:"attendance_id" binding:"required"`
		BreakType    string `json:"break_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attendance_id and break_type are required"})
		return
	}

	// Confirm attendance exists
	var attendance models.AttendanceLog
	if err := initializers.DB.First(&attendance, req.AttendanceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance log not found"})
		return
	}

	// Check if already on a break
	var activeBreak models.BreakLog
	if err := initializers.DB.
		Where("attendance_id = ? AND break_end IS NULL", req.AttendanceID).
		First(&activeBreak).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You must end your current break before starting a new one"})
		return
	}

	// Create new break
	newBreak := models.BreakLog{
		AttendanceID: req.AttendanceID,
		BreakType:    req.BreakType,
		BreakStart:   time.Now(),
	}

	if err := initializers.DB.Create(&newBreak).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start break"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Break started successfully",
		"break_id":    newBreak.ID,
		"break_start": newBreak.BreakStart,
	})
}


func EndBreak(c *gin.Context) {
	var req struct {
		AttendanceID uint `json:"attendance_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attendance_id is required"})
		return
	}

	// Check for open break
	var breakLog models.BreakLog
	if err := initializers.DB.
		Where("attendance_id = ? AND break_end IS NULL", req.AttendanceID).
		Order("created_at DESC").
		First(&breakLog).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active break found"})
		return
	}

	now := time.Now()
	breakLog.BreakEnd = &now

	if err := initializers.DB.Save(&breakLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end break"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Break ended successfully",
		"break_id":  breakLog.ID,
		"break_end": breakLog.BreakEnd,
	})
}


func GetEmployeeReports(c *gin.Context) {
	employeeID := c.Query("employee_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if employeeID == "" || startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id, start_date, and end_date are required"})
		return
	}

	var employee models.Employee
	if err := initializers.DB.First(&employee, employeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	var attendanceLogs []models.AttendanceLog
	if err := initializers.DB.
		Where("employee_id = ? AND DATE(clock_in) BETWEEN ? AND ?", employee.ID, startDate, endDate).
		Preload("Breaks").
		Find(&attendanceLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load attendance logs"})
		return
	}

	layout := "15:04"
	scheduledStart, _ := time.Parse(layout, employee.StartTime)
	reports := []gin.H{}

	for _, log := range attendanceLogs {
		if log.ClockOut == nil {
			continue // skip incomplete shifts
		}

		// Parse scheduled start for that day
		scheduled := time.Date(log.ClockIn.Year(), log.ClockIn.Month(), log.ClockIn.Day(),
			scheduledStart.Hour(), scheduledStart.Minute(), 0, 0, log.ClockIn.Location())

		lateMinutes := 0
		if log.ClockIn.After(scheduled) {
			lateMinutes = int(log.ClockIn.Sub(scheduled).Minutes())
		}

		breakSummary := map[string]struct {
			Duration int
			Count    int
		}{}
		totalBreakMinutes := 0

		for _, b := range log.Breaks {
			if b.BreakEnd == nil {
				continue
			}
			duration := int(b.BreakEnd.Sub(b.BreakStart).Minutes())
			summary := breakSummary[b.BreakType]
			summary.Duration += duration
			summary.Count++
			breakSummary[b.BreakType] = summary
			totalBreakMinutes += duration
		}

		workMinutes := int(log.ClockOut.Sub(log.ClockIn).Minutes()) - totalBreakMinutes
		workHours := float64(workMinutes) / 60.0
		breakHours := float64(totalBreakMinutes) / 60.0
		totalHours := float64(workMinutes+totalBreakMinutes) / 60.0
		totalWage := workHours * float64(employee.HourlyWage)

		breaks := []gin.H{}
		for breakType, info := range breakSummary {
			breaks = append(breaks, gin.H{
				"break_type":       breakType,
				"duration_minutes": info.Duration,
				"count":            info.Count,
			})
		}

		reports = append(reports, gin.H{
			"date":               log.ClockIn.Format("2006-01-02"),
			"clock_in":           log.ClockIn,
			"clock_out":          log.ClockOut,
			"breaks":             breaks,
			"total_worked_hours": workHours,
			"total_break_hours":  breakHours,
			"total_hours":        totalHours,
			"hourly_wage":        employee.HourlyWage,
			"total_wage":         totalWage,
			"late_minutes":       lateMinutes,
			"is_late":            lateMinutes > 0,
		})
	}

	c.JSON(http.StatusOK, reports)
}
