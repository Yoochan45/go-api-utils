package health

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// NewHandler returns a simple health-check handler that verifies DB connectivity.
// Example:
//
//	e := echo.New()
//	e.GET("/health", health.NewHandler(db))
func NewHandler(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type status struct {
			Status string `json:"status"`
			DB     string `json:"db"`
			Time   string `json:"time"`
		}
		dbStatus := "ok"
		if err := db.Exec("SELECT 1").Error; err != nil {
			dbStatus = "down"
		}
		return c.JSON(http.StatusOK, status{
			Status: "ok",
			DB:     dbStatus,
			Time:   time.Now().UTC().Format(time.RFC3339),
		})
	}
}
