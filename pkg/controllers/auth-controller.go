package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// @Summary Register a new employee
// @Description Creates a new employee account with a hashed password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param employee body models.EmployeeDTO true "Employee registration details"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Invalid input or creation failed"
// @Failure 500 {string} string "Failed to hash password"
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	var input models.Employee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	input.Password = string(hashedPassword)
	config.Connect()
	if err := config.GetDB().Create(&input).Error; err != nil {
		http.Error(w, "Failed to create employee", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful"})
}

// Login godoc
// @Summary Log in an employee
// @Description Authenticates employee by email and password and starts a session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body models.EmployeeDTO true "Login credentials (email and password)"
// @Success 200 {object} map[string]string "Login successful message"
// @Failure 400 {string} string "Invalid input"
// @Failure 401 {string} string "Email not found or incorrect password"
// @Router /login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	var input models.Employee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var employee models.Employee
	config.Connect()
	if err := config.GetDB().Where("email = ?", input.Email).First(&employee).Error; err != nil {
		http.Error(w, "Email not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(input.Password)); err != nil {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	sess, _ := session.GetStore().Get(r, "session")
	sess.Values["employee_id"] = employee.ID
	sess.Options.HttpOnly = true
	sess.Options.SameSite = http.SameSiteLaxMode
	sess.Options.Secure = false
	sess.Options.Path = "/"
	// if os.Getenv("ENV") != "production" {
	// 	sess.Options.Secure = false
	// } else {
	// 	sess.Options.Secure = true
	// }
	sess.Save(r, w)
	// fmt.Printf("Login success: session set for employee_id=%d\n", employee.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

// Logout godoc
// @Summary Log out the current employee
// @Description Ends the employee's session by clearing session data
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]string "Logout successful message"
// @Router /logout [post]
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := session.GetStore().Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}