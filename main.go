package main

import (
	"os"
	"time"

	"github.com/aoncodev/qrbackend/controllers"
	"github.com/aoncodev/qrbackend/initializers"
	"github.com/aoncodev/qrbackend/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)


func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}


func main() {
	r := gin.Default()

	// CORS configuration - allow both development and production origins
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "https://qrbackend-doo3.onrender.com", "https://www.qrbackend-doo3.onrender.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))


	r.POST("/api/admin/login", controllers.AdminLogin)
	r.POST("/api/employee/status", controllers.GetEmployeeStatus)
	r.POST("/api/employee/login", controllers.EmployeeLogin)
	r.GET("/api/employee/status/:id", controllers.GetEmployeeStatusByID)
	r.POST("/api/employee/clock-in", controllers.ClockIn)
	r.POST("/api/employee/clock-out", controllers.ClockOut)
	r.POST("/api/employee/break/start", controllers.StartBreak)
	r.POST("/api/employee/break/end", controllers.EndBreak)
	r.GET("/api/attendance/daily", controllers.GetDailyAttendance)



	admin := r.Group("/api")
	admin.Use(middleware.JWTAuthMiddleware())

	admin.GET("/employees", controllers.GetEmployees)
	admin.GET("/employees/:id", controllers.GetEmployeeByID) // Assuming this is for getting a specific employee
	admin.POST("/employees", controllers.CreateEmployee)
	admin.PUT("/employees/:id", controllers.UpdateEmployee)
	admin.DELETE("/employees/:id", controllers.DeleteEmployee)
	admin.GET("/employee/reports", controllers.GetEmployeeReports)

	// Attendance management endpoints
	admin.PUT("/attendance/:attendance_id", controllers.UpdateAttendance)
	admin.PUT("/attendance/:attendance_id/breaks", controllers.UpdateAttendanceBreaks)
	admin.POST("/attendance/:attendance_id/breaks", controllers.AddBreak)
	admin.DELETE("/attendance/:attendance_id/breaks/:break_id", controllers.DeleteBreak)

	r.Run(":8080") // listen and serve on localhost:8080
}
