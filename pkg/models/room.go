package models

import (
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	Name     string    `json:"name"`
	Capacity *int      `json:"capacity"`
	Location string    `json:"location"`
	Bookings []Booking `json:"bookings,omitempty"`
}

// RoomDTO represents a Room for Swagger
// swagger:model Room
type RoomDTO struct {
	Name     string `json:"name"`
	Capacity *int   `json:"capacity"`
	Location string `json:"location"`
}