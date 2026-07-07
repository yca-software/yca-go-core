package chi_server

import (
	"github.com/labstack/echo/v4"
	chi_error "github.com/yca-software/yca-go-core/error"
	chi_types "github.com/yca-software/yca-go-core/types"
)

// GetAccessInfo safely extracts the unified security context from the Echo context.
func GetAccessInfo(c echo.Context) (*chi_types.AccessInfo, error) {
	val := c.Get("accessInfo")
	if val == nil {
		return nil, chi_error.NewUnauthorizedError(nil, "", nil)
	}

	accessInfo, ok := val.(*chi_types.AccessInfo)
	if !ok || accessInfo == nil {
		return nil, chi_error.NewUnauthorizedError(nil, "", nil)
	}

	return accessInfo, nil
}
