package controllers

import (
	"fmt"
	"net/http"

	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/googleapi"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
)

// GoogleLogin godoc
// @Summary Initiate Google OAuth login flow
// @Description Redirects logged-in employee to Google OAuth consent screen for authentication
// @Tags Authentication
// @Produce plain
// @Success 307 "Redirect to Google OAuth consent screen"
// @Failure 401 {string} string "User not logged in"
// @Failure 500 {string} string "Failed to create auth URL"
// @Router /google/login [get]
func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.GetStore().Get(r, "session")
	userID, ok := sess.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	authURL, err := googleapi.GetAuthURLWithUser(userID)
	if err != nil {
		http.Error(w, "Failed to create auth URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallback godoc
// @Summary Handle Google OAuth callback
// @Description Processes OAuth code and state, exchanges code for tokens, and stores them linked to the employee
// @Tags Authentication
// @Produce plain
// @Param code query string true "OAuth authorization code"
// @Param state query string true "OAuth state parameter"
// @Success 200 {string} string "Google Calendar authorization successful! You may close this tab."
// @Failure 400 {string} string "Missing or invalid code/state parameter"
// @Failure 500 {string} string "Token exchange or database save failed"
// @Router /oauth2callback [get]
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing code in request", http.StatusBadRequest)
		return
	}
	if state == "" {
		http.Error(w, "Missing state in request", http.StatusBadRequest)
		return
	}

	userID, err := googleapi.ParseState(state)
	if err != nil {
		http.Error(w, "Invalid state parameter: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := googleapi.ExchangeCode(code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	newToken := models.GoogleToken{
		EmployeeID:   userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	db := config.GetDB()
	if err := db.Create(&newToken).Error; err != nil {
		http.Error(w, "Failed to save Google token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Google Calendar authorization successful! You may close this tab.")
}