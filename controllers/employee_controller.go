package controllers

import (
	"net/http"
	"time"

	"github.com/aoncodev/qrbackend/initializers"
	"github.com/aoncodev/qrbackend/models"
	"github.com/gin-gonic/gin"
)


func GetEmployees(c *gin.Context) {
	var employees []models.Employee
	if err := initializers.DB.Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employees"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"employees": employees})
}

func CreateEmployee(c *gin.Context) {
	var input models.Employee

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check required fields are not empty/zero
	if input.Name == "" || input.QRID == "" || input.HourlyWage == 0 || input.Role == "" || input.StartTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	if err := initializers.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var employee models.Employee

	if err := initializers.DB.First(&employee, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	if err := c.ShouldBindJSON(&employee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := initializers.DB.Save(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee"})
		return
	}

	c.JSON(http.StatusOK, employee)
}

func DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	if err := initializers.DB.Delete(&models.Employee{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete employee"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted"})
}

func GetEmployeeByID(c *gin.Context) {
	id := c.Param("id")
	var employee models.Employee

	if err := initializers.DB.First(&employee, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	c.JSON(http.StatusOK, employee)
}

// GetDailyAttendance returns daily attendance for all employees for a given date
func GetDailyAttendance(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date query param required (YYYY-MM-DD)"})
		return
	}
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	var employees []models.Employee
	if err := initializers.DB.Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employees"})
		return
	}

	results := []gin.H{}
	for _, emp := range employees {
		var attendance models.AttendanceLog
		loc, _ := time.LoadLocation("Asia/Seoul")
		kstDate, _ := time.ParseInLocation("2006-01-02", dateStr, loc)
		utcStart := kstDate.UTC()
		utcEnd := kstDate.Add(24 * time.Hour).UTC()

		err := initializers.DB.
			Where("employee_id = ? AND clock_in >= ? AND clock_in < ?", emp.ID, utcStart, utcEnd).
			Preload("Breaks").
			First(&attendance).Error
		if err != nil {
			results = append(results, gin.H{
				"employee_id": emp.ID,
				"employee": emp.Name,
				"clock_in": nil,
				"clock_out": nil,
				"total_hours": 0,
				"breaks": nil,
				"break_time": 0,
				"status": "absent",
			})
			continue
		}

		// Calculate total break time and check if currently on break
		totalBreakMinutes := 0
		breaks := []gin.H{}
		isOnBreak := false
		for _, b := range attendance.Breaks {
			if b.BreakEnd != nil {
				dur := int(b.BreakEnd.Sub(b.BreakStart).Minutes())
				breaks = append(breaks, gin.H{
					"break_type": b.BreakType,
					"start": b.BreakStart,
					"end": b.BreakEnd,
					"duration_minutes": dur,
				})
				totalBreakMinutes += dur
			} else {
				isOnBreak = true
				breaks = append(breaks, gin.H{
					"break_type": b.BreakType,
					"start": b.BreakStart,
					"end": nil,
					"duration_minutes": nil,
				})
			}
		}

		var totalHours float64
		var status string
		if attendance.ClockOut != nil {
			workMinutes := int(attendance.ClockOut.Sub(attendance.ClockIn).Minutes()) - totalBreakMinutes
			if workMinutes < 0 { workMinutes = 0 }
			totalHours = float64(workMinutes) / 60.0
			status = "present"
		} else if isOnBreak {
			workMinutes := int(time.Now().Sub(attendance.ClockIn).Minutes()) - totalBreakMinutes
			if workMinutes < 0 { workMinutes = 0 }
			totalHours = float64(workMinutes) / 60.0
			status = "on_break"
		} else {
			// Not clocked out yet and not on break
			workMinutes := int(time.Now().Sub(attendance.ClockIn).Minutes()) - totalBreakMinutes
			if workMinutes < 0 { workMinutes = 0 }
			totalHours = float64(workMinutes) / 60.0
			status = "working"
		}

		results = append(results, gin.H{
			"employee_id": emp.ID,
			"employee": emp.Name,
			"attendance_id": attendance.ID,
			"clock_in": attendance.ClockIn,
			"clock_out": attendance.ClockOut,
			"total_hours": totalHours,
			"breaks": breaks,
			"break_time": float64(totalBreakMinutes) / 60.0,
			"status": status,
		})
	}

	c.JSON(http.StatusOK, results)
}
