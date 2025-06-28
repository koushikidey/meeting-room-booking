package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"time"

	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/googleapi"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
	"github.com/koushikidey/go-meetingroombook/pkg/utils"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"gorm.io/gorm"
)

// CreateBooking godoc
// @Summary Create a new booking
// @Description Creates a new booking if the time and capacity constraints are satisfied. Sends confirmation email and adds Google Calendar event if linked.
// @Tags Bookings
// @Accept json
// @Produce json
// @Param booking body models.BookingDTO true "Booking request data"
// @Success 201 {object} models.BookingDTO
// @Failure 400 {string} string "Invalid input or time conflict"
// @Failure 401 {string} string "Unauthorized"
// @Failure 409 {string} string "Booking time conflict"
// @Failure 500 {string} string "Internal Server Error"
// @Router /bookings [post]
func CreateBooking(w http.ResponseWriter, r *http.Request) {
	sessionData, _ := session.GetStore().Get(r, "session")
	employeeID, ok := sessionData.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config.Connect()
	db := config.GetDB()

	var booking models.Booking
	if err := json.Unmarshal(body, &booking); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if booking.EndTime.Before(booking.StartTime) {
		http.Error(w, "End time is before start time", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateTimeFormat(booking.StartTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateTimeFormat(booking.EndTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingBookings []models.Booking
	db.Where("room_id = ?", booking.RoomID).Find(&existingBookings)

	var room models.Room
	db.Where("ID = ?", booking.RoomID).Find(&room)
	numberOfAttendees := booking.NumAttendees
	maxCapacity := *room.Capacity
	if _, err := utils.IsCapacityExceeding(numberOfAttendees, maxCapacity); err != nil {
		http.Error(w, "Capacity Exceeded", http.StatusBadRequest)
		return
	}

	for _, b := range existingBookings {
		conflict, err := utils.IsBookingConflict(booking.StartTime, booking.EndTime, b.StartTime, b.EndTime, b.Room, booking.Room)
		if err != nil {
			http.Error(w, "Error checking for conflicts", http.StatusInternalServerError)
			return
		}
		if conflict {
			http.Error(w, "Booking time conflicts with an existing booking", http.StatusConflict)
			return
		}
	}

	booking.EmployeeID = employeeID
	if err := db.Create(&booking).Error; err != nil {
		http.Error(w, "Could not create booking", http.StatusInternalServerError)
		return
	}

	var employee models.Employee
	db.First(&employee, employeeID)
	message := fmt.Sprintf("Hi %s,\n\nYour meeting room booking is confirmed from %s to %s in Room ID %d.",
		employee.Name, booking.StartTime, booking.EndTime, booking.RoomID)
	go utils.SendEmail(employee.Email, "Meeting Room Booking Confirmation", message)

	var token models.GoogleToken
	if err := db.Where("employee_id = ?", employeeID).First(&token).Error; err == nil {
		oauthToken := &oauth2.Token{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Expiry:       token.Expiry,
		}

		client := googleapi.GetClient(oauthToken)

		srv, err := calendar.New(client)
		if err == nil {
			event := &calendar.Event{
				Summary:     "Meeting Room Booking",
				Location:    fmt.Sprintf("Room ID %d", booking.RoomID),
				Description: fmt.Sprintf("Booked by %s", employee.Name),
				Start: &calendar.EventDateTime{
					DateTime: booking.StartTime.Format(time.RFC3339),
					TimeZone: "Asia/Kolkata",
				},
				End: &calendar.EventDateTime{
					DateTime: booking.EndTime.Format(time.RFC3339),
					TimeZone: "Asia/Kolkata",
				},
			}

			createdEvent, err := srv.Events.Insert("primary", event).Do()
			if err != nil {
				fmt.Printf("Failed to create Google Calendar event: %v\n", err)
			} else {

				booking.CalendarID = createdEvent.Id
				result := db.Model(&models.Booking{}).Where("id = ?", booking.ID).Update("calendar_id", createdEvent.Id)
				if result.Error != nil {
					fmt.Println("Failed to update calendar ID:", result.Error)
				}

			}

		} else {
			fmt.Printf("Failed to create Google Calendar client: %v\n", err)
		}
	} else {
		fmt.Println("Google Calendar not linked for employee:", employeeID)
	}

	resp, _ := json.Marshal(booking)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// GetBookings godoc
// @Summary Get list of all bookings
// @Description Retrieves all bookings along with all details
// @Tags Bookings
// @Produce json
// @Success 200 {array} models.BookingDTO
// @Failure 500 {string} string "Internal Server Error"
// @Router /bookings [get]
func GetBookings(w http.ResponseWriter, r *http.Request) {
	// session, _ := session.GetStore().Get(r, "session")
	// employeeID, ok := session.Values["employee_id"].(uint)
	// if !ok {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	var bookings []models.Booking
	config.Connect()
	db := config.GetDB()
	//db.Preload("Room").Preload("Employee").Where("employee_id = ?", employeeID).Find(&bookings)
	db.Preload("Room").Preload("Employee").Find(&bookings)
	resp, _ := json.Marshal(bookings)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// GetBooking godoc
// @Summary Get booking by booking ID
// @Description Retrieves booking details by ID including employee and room info
// @Tags Bookings
// @Produce json
// @Param id path uint true "Booking ID"
// @Success 200 {object} models.BookingDTO
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Booking not found"
// @Failure 500 {string} string "Error marshalling data"
// @Router /bookings/{id} [get]
func GetBooking(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var booking models.Booking
	config.Connect()
	db := config.GetDB()
	result := db.Preload("Room").Preload("Employee").First(&booking, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Booking not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve booking", http.StatusInternalServerError)
		return
	}

	if booking.EmployeeID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	resp, _ := json.Marshal(booking)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// UpdateBooking godoc
// @Summary Update existing booking details
// @Description Allows an authenticated employee to update their booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param booking body models.BookingDTO true "Updated booking details"
// @Success 200 {object} models.BookingDTO
// @Failure 400 {string} string "Invalid Employee ID or JSON input"
// @Failure 401 {string} string "Unauthorized (not logged in)"
// @Failure 403 {string} string "Forbidden (trying to update another employee's booking)"
// @Failure 404 {string} string "Booking not found"
// @Failure 500 {string} string "Failed to update booking"
// @Router /booking/{id} [put]
func UpdateBooking(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var existing models.Booking
	config.Connect()
	db := config.GetDB()
	if err := db.First(&existing, id).Error; err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	if existing.EmployeeID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var updated models.Booking
	if err := json.Unmarshal(body, &updated); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateTimeFormat(updated.StartTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateTimeFormat(updated.EndTime); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var existingBookings []models.Booking
	db.Where("room_id = ?", updated.RoomID).Find(&existingBookings)

	var room models.Room
	db.Where("ID = ?", updated.RoomID).Find(&room)
	numberOfAttendees := updated.NumAttendees
	maxCapacity := *room.Capacity
	if _, err := utils.IsCapacityExceeding(numberOfAttendees, maxCapacity); err != nil {
		http.Error(w, "Capacity Exceeded", http.StatusBadRequest)
		return
	}

	var conflicts []models.Booking
	db.Where("room_id = ? AND id != ?", updated.RoomID, id).Find(&conflicts)
	for _, b := range conflicts {
		conflict, err := utils.IsBookingConflict(updated.StartTime, updated.EndTime, b.StartTime, b.EndTime, updated.Room, b.Room)
		if err != nil {
			http.Error(w, "Error checking for conflicts", http.StatusInternalServerError)
			return
		}
		if conflict {
			http.Error(w, "Updated time conflicts with another booking", http.StatusConflict)
			return
		}
	}
	err = googleapi.DeleteCalendarEvent(existing.EmployeeID, existing.CalendarID)
	if err != nil {
		log.Println("Failed to delete calendar event:", err)
	}

	existing.RoomID = updated.RoomID
	existing.EmployeeID = updated.EmployeeID
	existing.StartTime = updated.StartTime
	existing.EndTime = updated.EndTime
	existing.NumAttendees = updated.NumAttendees

	if err := db.Save(&existing).Error; err != nil {
		http.Error(w, "Failed to update booking", http.StatusInternalServerError)
		return
	}

	var employee models.Employee
	db.First(&employee, employeeID)
	message := fmt.Sprintf("Hi %s,\n\nYour meeting room booking is confirmed from %s to %s in Room ID %d.",
		employee.Name, existing.StartTime, existing.EndTime, existing.RoomID)
	go utils.SendEmail(employee.Email, "Meeting Room Booking Updated and Confirmed", message)

	var token models.GoogleToken
	if err := db.Where("employee_id = ?", employeeID).First(&token).Error; err == nil {
		oauthToken := &oauth2.Token{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Expiry:       token.Expiry,
		}

		client := googleapi.GetClient(oauthToken)

		srv, err := calendar.New(client)
		if err == nil {
			event := &calendar.Event{
				Summary:     "Meeting Room Booking",
				Location:    fmt.Sprintf("Room ID %d", existing.RoomID),
				Description: fmt.Sprintf("Booked by %s", employee.Name),
				Start: &calendar.EventDateTime{
					DateTime: existing.StartTime.Format(time.RFC3339),
					TimeZone: "Asia/Kolkata",
				},
				End: &calendar.EventDateTime{
					DateTime: existing.EndTime.Format(time.RFC3339),
					TimeZone: "Asia/Kolkata",
				},
			}
			createdEvent, err := srv.Events.Insert("primary", event).Do()
			if err != nil {
				fmt.Printf("Failed to create Google Calendar event: %v\n", err)
			} else {

				existing.CalendarID = createdEvent.Id
				result := db.Model(&models.Booking{}).Where("id = ?", existing.ID).Update("calendar_id", createdEvent.Id)
				if result.Error != nil {
					fmt.Println("Failed to update calendar ID:", result.Error)
				}

			}

		} else {
			fmt.Printf("Failed to create Google Calendar client: %v\n", err)
		}
	} else {
		fmt.Println("Google Calendar not linked for employee:", employeeID)
	}

	resp, _ := json.Marshal(existing)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// DeleteBooking godoc
// @Summary Delete existing booking details
// @Description Allows an authenticated employee to delete their booking
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param booking body models.BookingDTO true "Deleted booking details"
// @Failure 400 {string} string "Invalid Employee ID or JSON input"
// @Failure 401 {string} string "Unauthorized (not logged in)"
// @Failure 403 {string} string "Forbidden (trying to update another employee's booking)"
// @Failure 404 {string} string "Booking not found"
// @Failure 500 {string} string "Failed to delete booking"
// @Router /booking/{id} [delete]
func DeleteBooking(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}
	config.Connect()
	db := config.GetDB()
	var booking models.Booking
	if err := db.First(&booking, id).Error; err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	if booking.EmployeeID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := db.Delete(&booking).Error; err != nil {
		http.Error(w, "Failed to delete booking", http.StatusInternalServerError)
		return
	}

	err = googleapi.DeleteCalendarEvent(booking.EmployeeID, booking.CalendarID)
	if err != nil {
		log.Println("Failed to delete calendar event:", err)
	}
	var employee models.Employee
	db.First(&employee, employeeID)
	message := fmt.Sprintf("Hi %s,\n\nYour meeting room booking which was confirmed from %s to %s in Room ID %d has been deleted.",
		employee.Name, booking.StartTime, booking.EndTime, booking.RoomID)
	go utils.SendEmail(employee.Email, "Meeting Room Booking Cancelled", message)

	w.WriteHeader(http.StatusNoContent)
}