package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  string
		expectedCode int
	}{
		{
			name: "Valid request body",
			requestBody: `{
				"email": "test@gmail.com",
				"password": "password123"
			}`,
			expectedCode: http.StatusCreated,
		},
		{
			name: "Invalid request body",
			requestBody: `{
				"email": "test@gmail.com"
			}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/signup", CreateAccount)

			req, err := http.NewRequest("POST", "/signup", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// func TestCreateAccount(t *testing.T) {

// }
