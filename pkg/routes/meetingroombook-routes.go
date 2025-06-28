package routes

import (
	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/controllers"
)

func RegisterMeetingRoomRoutes(router *mux.Router) {
	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/logout", controllers.Logout).Methods("POST")

	router.HandleFunc("/employees", controllers.GetEmployees).Methods("GET")
	router.HandleFunc("/employees/{id}", controllers.GetEmployeeByIDWithCache).Methods("GET")
	//router.HandleFunc("/employees/{id}", controllers.GetEmployee).Methods("GET")
	router.HandleFunc("/employees/{id}", controllers.UpdateEmployees).Methods("PUT")

	router.HandleFunc("/rooms", controllers.CreateRoom).Methods("POST")
	router.HandleFunc("/rooms", controllers.GetRooms).Methods("GET")
	router.HandleFunc("/rooms/{id}", controllers.UpdateRoom).Methods("PUT")

	router.HandleFunc("/bookings", controllers.CreateBooking).Methods("POST")
	router.HandleFunc("/bookings", controllers.GetBookings).Methods("GET")
	router.HandleFunc("/bookings/{id}", controllers.GetBooking).Methods("GET")
	router.HandleFunc("/bookings/{id}", controllers.UpdateBooking).Methods("PUT")
	router.HandleFunc("/bookings/{id}", controllers.DeleteBooking).Methods("DELETE")

	router.HandleFunc("/google/login", controllers.GoogleLogin).Methods("GET")
	router.HandleFunc("/oauth2callback", controllers.GoogleCallback).Methods("GET")

}