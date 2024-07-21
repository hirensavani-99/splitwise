package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mocking the Wallet model and DB
type MockWallet struct{}

func (w *MockWallet) Get(db interface{}, userId int64) error {
	if userId == 1 {
		return nil
	}
	return fmt.Errorf("wallet not found")
}

func TestGetWalletById(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userId       string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Valid userId",
			userId:       "1",
			expectedCode: http.StatusOK,
			expectedBody: `{"wallet":{}}`,
		},
		{
			name:         "Invalid userId format",
			userId:       "abc",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"message":"Invalid userId"}`,
		},
		{
			name:         "Wallet not found",
			userId:       "2",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"message":"issue returning wallet","err":"wallet not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.GET("/wallet/:userId", func(c *gin.Context) {
				userId, err := strconv.ParseInt(c.Param("userId"), 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid userId"})
					return
				}
				wallet := &MockWallet{}
				err = wallet.Get(nil, userId)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "issue returning wallet", "err": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"wallet": wallet})
			})

			req, _ := http.NewRequest(http.MethodGet, "/wallet/"+tt.userId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
