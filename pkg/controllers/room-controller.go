package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	"github.com/koushikidey/go-meetingroombook/pkg/utils"
	"gorm.io/gorm"
)

// func CreateRoom(w http.ResponseWriter, r *http.Request) {
// 	var room models.Room

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Failed to read request body", http.StatusBadRequest)
// 		return
// 	}

// 	if err := json.Unmarshal(body, &room); err != nil {
// 		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	config.Connect()
// 	db := config.GetDB()
// 	if err := db.Create(&room).Error; err != nil {
// 		http.Error(w, "Could not create room: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	var createdRoom models.Room
// 	if err := db.First(&createdRoom, room.ID).Error; err != nil {
// 		http.Error(w, "Could not retrieve created room "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(createdRoom)
// }

// func GetRooms(w http.ResponseWriter, r *http.Request) {
// 	var rooms []models.RoomDTO
// 	config.Connect()
// 	db := config.GetDB()
// 	db.Preload("Bookings.Room").Preload("Bookings.Employee").Find(&rooms)

//		resp, _ := json.Marshal(rooms)
//		w.Header().Set("Content-type", "application/json")
//		w.Write(resp)
//	}
func GetRoomsWithDB(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rooms []models.Room
		db.Preload("Bookings.Room").Preload("Bookings.Employee").Find(&rooms)

		resp, _ := json.Marshal(rooms)
		w.Header().Set("Content-type", "application/json")
		w.Write(resp)
	}
}

// GetRooms godoc
// @Summary Get list of all rooms
// @Description Retrieves all rooms along with their bookings and booked employees
// @Tags Rooms
// @Produce json
// @Success 200 {array} models.RoomDTO
// @Failure 500 {string} string "Internal Server Error"
// @Router /rooms [get]
func GetRooms(w http.ResponseWriter, r *http.Request) {
	//config.Connect()
	db := config.GetDB()
	GetRoomsWithDB(db)(w, r)
}

func CreateRoomWithDB(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var room models.Room
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &room); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if err := db.Create(&room).Error; err != nil {
			http.Error(w, "Could not create room: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var createdRoom models.Room
		if err := db.First(&createdRoom, room.ID).Error; err != nil {
			http.Error(w, "Could not retrieve created room "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdRoom)

	}
}

// CreateRoom godoc
// @Summary Create a new room
// @Description Adds a new meeting room to the system
// @Tags Rooms
// @Accept json
// @Produce json
// @Param room body models.RoomDTO true "Room details"
// @Success 201 {object} models.RoomDTO
// @Failure 400 {string} string "Invalid JSON or bad request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /rooms [post]
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	//config.Connect()
	db := config.GetDB()
	CreateRoomWithDB(db)(w, r)
}

// UpdateRoom godoc
// @Summary Update room details
// @Description Update details of existing room by ID
// @Tags Rooms
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param room body models.RoomDTO true "Room details to update"
// @Success 200 {object} models.RoomDTO
// @Failure 400 {string} string "Invalid JSON or bad request"
// @Failure 404 {string} string "Room not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /rooms/{id} [put]
func UpdateRoom(w http.ResponseWriter, r *http.Request) {
	var updateRoom = &models.Room{}
	utils.ParseBody(r, updateRoom)
	vars := mux.Vars(r)
	room_id := vars["id"]
	ID, err := strconv.ParseInt(room_id, 0, 0)
	if err != nil {
		fmt.Println("Error while parsing")
	}
	config.Connect()
	db := config.GetDB()
	var getRoom models.Room
	db = db.Where("ID=?", ID).Find(&getRoom)
	if updateRoom.Name != "" {
		getRoom.Name = updateRoom.Name
	}
	if updateRoom.Location != "" {
		getRoom.Location = updateRoom.Location
	}
	if updateRoom.Capacity != nil {
		getRoom.Capacity = updateRoom.Capacity
	}

	db.Save(&getRoom)
	res, _ := json.Marshal(&getRoom)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}