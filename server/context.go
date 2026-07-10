package yca_server

import (
	"github.com/labstack/echo/v4"
	yca_error "github.com/yca-software/yca-go-core/error"
	yca_types "github.com/yca-software/yca-go-core/types"
)

// GetAccessInfo safely extracts the unified security context from the Echo context.
func GetAccessInfo(c echo.Context) (*yca_types.AccessInfo, error) {
	val := c.Get("accessInfo")
	if val == nil {
		return nil, yca_error.NewUnauthorizedError(nil, "", nil)
	}

	accessInfo, ok := val.(*yca_types.AccessInfo)
	if !ok || accessInfo == nil {
		return nil, yca_error.NewUnauthorizedError(nil, "", nil)
	}

	return accessInfo, nil
}
