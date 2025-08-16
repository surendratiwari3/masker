package masker

import (
	"github.com/labstack/echo/v4"
)

// JSON masks v before sending JSON response.
func JSON(c echo.Context, code int, v interface{}) error {
	Mask(v)
	return c.JSON(code, v)
}

