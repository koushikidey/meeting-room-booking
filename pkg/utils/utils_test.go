package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCapacityExceeding(t *testing.T) {
	tests := []struct {
		name          string
		NumAttendees  int
		maxCapacity   int
		boolExpected  bool
		expectedError error
	}{
		{
			name:          "Capacity within limit",
			NumAttendees:  20,
			maxCapacity:   80,
			boolExpected:  false,
			expectedError: nil,
		},
		{
			name:          "Capacity equals limit",
			NumAttendees:  80,
			maxCapacity:   80,
			boolExpected:  false,
			expectedError: nil,
		},
		{
			name:          "Capacity exceeds limit",
			NumAttendees:  81,
			maxCapacity:   80,
			boolExpected:  true,
			expectedError: errors.New("capacity exceeded"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := IsCapacityExceeding(test.NumAttendees, test.maxCapacity)
			assert.Equal(t, test.boolExpected, result)
			if test.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}