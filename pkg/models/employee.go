package models

import (
	"gorm.io/gorm"
)

type Employee struct {
	gorm.Model
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Bookings []Booking `json:"bookings,omitempty"`
}

// EmployeeDTO represents an employee for Swagger
// swagger:model Employee
type EmployeeDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}