package status

import (
	"fmt"
	"net/http"
)

var (
	SuccEmailAddedToMailingList = fmt.Sprintf("your email was added to the mailing list")
)

func SuccessStatusOK(msg string) Toast {
	return Toast{
		Message:    msg,
		StatusCode: http.StatusOK,
	}
}

func SuccessStatusCreated(msg string) Toast {
	return Toast{
		Message:    msg,
		StatusCode: http.StatusCreated,
	}
}

func SuccessStatusNoContent(msg string) Toast {
	return Toast{
		Message:    msg,
		StatusCode: http.StatusNoContent,
	}
}
