package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/koushikidey/go-meetingroombook/pkg/models"
)

func ValidateTimeFormat(t time.Time) error {
	if t.IsZero() {
		return errors.New("time cannot be zero value")
	}

	return nil
}

func IsBookingConflict(start1, end1, start2, end2 time.Time, room1, room2 models.Room) (bool, error) {
	if start1.IsZero() || end1.IsZero() || start2.IsZero() || end2.IsZero() {
		return false, errors.New("one or more times are zero")
	}

	if start1.Before(end2) && start2.Before(end1) && room1.ID == room2.ID {
		return true, nil
	}
	return false, nil
}

func ParseBody(r *http.Request, x interface{}) {
	if body, err := io.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal([]byte(body), x); err != nil {
			return
		}
	}
}

var ErrCapacityExceeded = errors.New("capacity exceeded")

func IsCapacityExceeding(numberOfAttendees, maxCapacity int) (bool, error) {
	if numberOfAttendees > maxCapacity {
		return true, ErrCapacityExceeded
	}
	return false, nil
}