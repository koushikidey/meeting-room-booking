package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/cache"
	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// func CreateEmployee(w http.ResponseWriter, r *http.Request) {
// 	var emp models.Employee

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Failed to read request body", http.StatusBadRequest)
// 		return
// 	}

// 	if err := json.Unmarshal(body, &emp); err != nil {
// 		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	config.Connect()
// 	db := config.GetDB()
// 	if err := db.Create(&emp).Error; err != nil {
// 		http.Error(w, "Could not create employee: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	var createdEmployee models.Employee
// 	if err := db.First(&createdEmployee, emp.ID).Error; err != nil {
// 		http.Error(w, "Could not retrieve created employee: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(createdEmployee)
// }

// GetEmployees godoc
// @Summary Get all employees
// @Description Returns a list of all employees with their bookings and room details
// @Tags Employees
// @Produce json
// @Success 200 {array} models.EmployeeDTO
// @Router /employees [get]
func GetEmployees(w http.ResponseWriter, r *http.Request) {
	var employees []models.Employee
	config.Connect()
	db := config.GetDB()
	db.Preload("Bookings.Room").Preload("Bookings.Employee").Find(&employees)
	resp, _ := json.Marshal(employees)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// GetEmployeeByIDWithCache godoc
// @Summary Get employee by ID (with cache)
// @Description Retrieves employee details by ID including bookings and room info, uses cache for faster response
// @Tags Employees
// @Produce json
// @Param id path uint true "Employee ID"
// @Success 200 {object} models.EmployeeDTO
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Employee not found"
// @Failure 500 {string} string "Error marshalling data"
// @Router /employees/{id} [get]
func GetEmployeeByIDWithCache(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars["id"]

	idUint, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if res, ok := cache.C.Read(uint(idUint)); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		return
	}

	db := config.GetDB()
	var employee models.Employee
	result := db.Preload("Bookings.Room").Preload("Bookings.Employee").First(&employee, idUint)
	if result.Error != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(employee)
	if err != nil {
		http.Error(w, "Error marshalling", http.StatusInternalServerError)
		return
	}

	cache.C.Update(uint(idUint), employee)

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// UpdateEmployees godoc
// @Summary Update an employee's own details
// @Description Allows an authenticated employee to update their name, email, and password
// @Tags Employees
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Param employee body models.EmployeeDTO true "Updated employee details (name, email, password)"
// @Success 200 {object} models.EmployeeDTO
// @Failure 400 {string} string "Invalid Employee ID or JSON input"
// @Failure 401 {string} string "Unauthorized (not logged in)"
// @Failure 403 {string} string "Forbidden (trying to update another employee)"
// @Failure 404 {string} string "Employee not found"
// @Failure 500 {string} string "Failed to update employee"
// @Router /employees/{id} [put]
func UpdateEmployees(w http.ResponseWriter, r *http.Request) {

	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid Employee ID", http.StatusBadRequest)
		return
	}

	var existing models.Employee
	config.Connect()
	db := config.GetDB()
	if err := db.First(&existing, id).Error; err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	if existing.ID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var updated models.Employee
	if err := json.Unmarshal(body, &updated); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	existing.Name = updated.Name
	existing.Email = updated.Email
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updated.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	existing.Password = string(hashedPassword)

	if err := db.Save(&existing).Error; err != nil {
		http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		return
	}
	resp, _ := json.Marshal(existing)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

}

func GetEmployee(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	employeeID, ok := session.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid Employee id", http.StatusBadRequest)
		return
	}

	var employee models.Employee
	config.Connect()
	db := config.GetDB()
	result := db.Preload("Bookings.Room").Preload("Bookings.Employee").First(&employee, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve employee", http.StatusInternalServerError)
		return
	}

	if employee.ID != employeeID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	resp, _ := json.Marshal(employee)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}