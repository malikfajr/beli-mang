package exception

import "net/http"

type CustomError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

// implement error interface
func (c *CustomError) Error() string {
	return c.Message
}

func Conflict(msg string) *CustomError {
	return &CustomError{
		Message:    msg,
		StatusCode: http.StatusConflict,
	}
}

func NotFound(msg string) *CustomError {
	return &CustomError{
		Message:    msg,
		StatusCode: http.StatusNotFound,
	}
}

func BadRequest(msg string) *CustomError {
	return &CustomError{
		Message:    msg,
		StatusCode: http.StatusBadRequest,
	}
}

func Unauthorized(msg string) *CustomError {
	return &CustomError{
		Message:    msg,
		StatusCode: http.StatusUnauthorized,
	}
}

func ServerError(msg string) *CustomError {
	return &CustomError{
		Message:    msg,
		StatusCode: http.StatusInternalServerError,
	}
}
