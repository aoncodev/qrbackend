package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aoncodev/qrbackend/initializers"
	"github.com/aoncodev/qrbackend/models"
	"github.com/gin-gonic/gin"
)

// UpdateAttendance updates an attendance record
func UpdateAttendance(c *gin.Context) {
	attendanceID := c.Param("attendance_id")
	id, err := strconv.ParseUint(attendanceID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendance ID"})
		return
	}

	var req struct {
		ClockIn  *time.Time `json:"clock_in"`
		ClockOut *time.Time `json:"clock_out"`
		Status   string     `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var attendance models.AttendanceLog
	if err := initializers.DB.First(&attendance, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	// Update fields if provided
	if req.ClockIn != nil {
		attendance.ClockIn = *req.ClockIn
	}
	if req.ClockOut != nil {
		attendance.ClockOut = req.ClockOut
	}

	if err := initializers.DB.Save(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Attendance updated successfully",
		"attendance": attendance,
	})
}

// UpdateAttendanceBreaks updates all breaks for an attendance record
func UpdateAttendanceBreaks(c *gin.Context) {
	attendanceID := c.Param("attendance_id")
	id, err := strconv.ParseUint(attendanceID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendance ID"})
		return
	}

	var req struct {
		Breaks []struct {
			BreakType string     `json:"break_type"`
			Start     time.Time  `json:"start"`
			End       *time.Time `json:"end"`
		} `json:"breaks"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var attendance models.AttendanceLog
	if err := initializers.DB.First(&attendance, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	// Delete existing breaks
	if err := initializers.DB.Where("attendance_id = ?", id).Delete(&models.BreakLog{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing breaks"})
		return
	}

	// Create new breaks
	var breaks []models.BreakLog
	for _, b := range req.Breaks {
		breakLog := models.BreakLog{
			AttendanceID: uint(id),
			BreakType:    b.BreakType,
			BreakStart:   b.Start,
			BreakEnd:     b.End,
		}
		breaks = append(breaks, breakLog)
	}

	if len(breaks) > 0 {
		if err := initializers.DB.Create(&breaks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create breaks"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Breaks updated successfully",
		"breaks":  breaks,
	})
}

// AddBreak adds a new break to an attendance record
func AddBreak(c *gin.Context) {
	attendanceID := c.Param("attendance_id")
	id, err := strconv.ParseUint(attendanceID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendance ID"})
		return
	}

	var req struct {
		BreakType string     `json:"break_type"`
		Start     time.Time  `json:"start"`
		End       *time.Time `json:"end"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var attendance models.AttendanceLog
	if err := initializers.DB.First(&attendance, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	breakLog := models.BreakLog{
		AttendanceID: uint(id),
		BreakType:    req.BreakType,
		BreakStart:   req.Start,
		BreakEnd:     req.End,
	}

	if err := initializers.DB.Create(&breakLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create break"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Break added successfully",
		"break":   breakLog,
	})
}

// DeleteBreak deletes a specific break
func DeleteBreak(c *gin.Context) {
	attendanceID := c.Param("attendance_id")
	breakID := c.Param("break_id")

	attID, err := strconv.ParseUint(attendanceID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendance ID"})
		return
	}

	breakIDUint, err := strconv.ParseUint(breakID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid break ID"})
		return
	}

	// Check if attendance exists
	var attendance models.AttendanceLog
	if err := initializers.DB.First(&attendance, attID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	// Delete the break
	if err := initializers.DB.Where("id = ? AND attendance_id = ?", breakIDUint, attID).Delete(&models.BreakLog{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete break"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Break deleted successfully",
	})
} 