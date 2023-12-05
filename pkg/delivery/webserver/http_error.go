package webserver

import "fmt"

type HttpError struct {
	StatusCode int
	Message    string
}

func (o *HttpError) Error() string {
	return fmt.Sprintf("http status code: %d, %s", o.StatusCode, o.Message)
}
