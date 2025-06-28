package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"
	"io"

	// "log"
	// "os"

	// "github.com/joho/godotenv"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBforGet() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}
	db.AutoMigrate(&models.Room{}, &models.Booking{}, &models.Employee{})
	cap := 10
	room := models.Room{
		//ID:       1,
		Name:     "Test Room",
		Location: "Test Location",
		Capacity: &cap,
	}
	db.Create(&room)
	return db
}

func TestGetRoomsWithDB(t *testing.T) {
	db := setupTestDBforGet()
	req, err := http.NewRequest("GET", "/rooms", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := GetRoomsWithDB(db)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response []models.Room
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "Test Room", response[0].Name)
}

// func TestMain(m *testing.M) {
// 	err := godotenv.Load(".env.test")
// 	if err != nil {
// 		log.Println("Warning: .env.test file not found")
// 	}

// 	code := m.Run()
// 	os.Exit(code)
// }

func setupTestDBforCreate() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}
	db.AutoMigrate(&models.Room{}, &models.Booking{}, &models.Employee{})
	return db
}

func TestCreateRoomWithDB(t *testing.T) {
	db := setupTestDBforCreate()
	roomData := models.Room{
		Name:     "Test Room",
		Location: "Test Location",
	}
	capacity := 50
	roomData.Capacity = &capacity
	roomJSON, err := json.Marshal(roomData)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/rooms", bytes.NewBuffer(roomJSON))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := CreateRoomWithDB(db)
	handler.ServeHTTP(rr, req)

	var createdRoom models.Room
	body, err := io.ReadAll(rr.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(body, &createdRoom)
	assert.NoError(t, err)

	assert.Equal(t, roomData.Name, createdRoom.Name)
	assert.Equal(t, roomData.Location, createdRoom.Location)
	assert.NotNil(t, createdRoom.Capacity)
	assert.Equal(t, *roomData.Capacity, *createdRoom.Capacity)

}