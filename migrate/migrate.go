package main

import (
	"github.com/aoncodev/qrbackend/initializers"
	"github.com/aoncodev/qrbackend/models"
)

func init() {
	// Load environment variables
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}

func main() {
	initializers.DB.AutoMigrate(
		&models.Employee{},
		&models.AttendanceLog{},
		&models.BreakLog{},
	)
}
