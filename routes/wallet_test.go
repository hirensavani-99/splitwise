package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"hirensavani.com/db"
)

func TestGetWalletById(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Invalid userId", func(t *testing.T) {
		r := gin.Default()
		r.GET("/getWallet/:userId", getWalletById)

		req, _ := http.NewRequest(http.MethodGet, "/getWallet/invalid_id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"message": "Invalid userId"}`, w.Body.String())
	})

	t.Run("Database error on Get", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer dbMock.Close()

		db.DB = dbMock

		// Simulate an error from the database
		mock.ExpectQuery("SELECT * FROM wallets WHERE user_id = \\$1").
			WithArgs(int64(1)).
			WillReturnError(assert.AnError)

		r := gin.Default()
		r.GET("/getWallet/:userId", getWalletById)

		req, _ := http.NewRequest(http.MethodGet, "/getWallet/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"message":"issue returning wallet"`)
	})

	// t.Run("Successful retrieval of wallet", func(t *testing.T) {

	// 	dbMock, mock, err := sqlmock.New()
	// 	assert.NoError(t, err)
	// 	defer dbMock.Close()

	// 	db.DB = dbMock

	// 	// Create rows without CreatedAt and UpdatedAt fields
	// 	rows := sqlmock.NewRows([]string{"user_id", "balance", "currency"}).
	// 		AddRow(1, 500.00, "USD")

	// 	// Adjust the query pattern to match the fields being selected
	// 	mock.ExpectQuery(`SELECT user_id, balance, currency FROM wallets WHERE user_id = \$1`).
	// 		WithArgs(int64(1)).
	// 		WillReturnRows(rows)

	// 	r := gin.Default()
	// 	r.GET("/getWallet/:userId", getWalletById)

	// 	req, err := http.NewRequest(http.MethodGet, "/getWallet/1", nil)

	// 	fmt.Println(err)
	// 	if err != nil {
	// 		t.Fatal("err-> %w", err)
	// 	}
	// 	w := httptest.NewRecorder()
	// 	r.ServeHTTP(w, req)

	// 	assert.Equal(t, http.StatusOK, w.Code)

	// 	// Print actual response for debugging
	// 	fmt.Println("---> Actual Response:", w.Body.String())

	// 	// Update expected JSON to match the simplified response
	// 	expected := `{
	//         "wallet": {
	//             "UserID": 1,
	//             "Balance": 500.00,
	//             "Currency": "USD"
	//         }
	//     }`
	// 	assert.JSONEq(t, expected, w.Body.String())
	// })
}
