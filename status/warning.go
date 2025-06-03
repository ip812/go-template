package status

import (
	"fmt"
	"net/http"
)

var (
	WarnEmailIsRequred     = fmt.Errorf("email is required")
	WarnInvalidEmailFormat = fmt.Errorf("provided email is not in valid format")
	WarnEmailAlreadyExists = fmt.Errorf("provided email already exists")
)

func WarningStatusBadRequest(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusBadRequest,
	}
}

func WarningStatunUnauthorized(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusUnauthorized,
	}
}

func WarningStatusForbidden(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusForbidden,
	}
}
