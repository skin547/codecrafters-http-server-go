package internal

import "fmt"

type NotFoundError struct {
	FileName string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("file not found: %s", e.FileName)
}

type InternalServerError struct {
	Reason string
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("internal server error: %s", e.Reason)
}
