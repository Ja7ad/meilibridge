package bridge

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_IsValidToken(t *testing.T) {
	tests := []struct {
		Name    string
		Request func() *http.Request
		Token   string
		Valid   bool
	}{
		{
			Name: "valid",
			Request: func() *http.Request {
				r := httptest.NewRequest("POST", "/", nil)
				r.Header.Add(_triggerHeaderKey, "foobar")
				return r
			},
			Token: "foobar",
			Valid: true,
		},
		{
			Name: "empty token",
			Request: func() *http.Request {
				r := httptest.NewRequest("POST", "/", nil)
				return r
			},
			Token: "",
			Valid: true,
		},
		{
			Name: "invalid token",
			Request: func() *http.Request {
				r := httptest.NewRequest("POST", "/", nil)
				r.Header.Add(_triggerHeaderKey, "foobar")
				return r
			},
			Token: "bar",
			Valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r := tt.Request()
			if tt.Valid {
				assert.True(t, isValidTriggerToken(r, tt.Token))
			} else {
				assert.False(t, isValidTriggerToken(r, tt.Token))
			}
		})
	}
}
