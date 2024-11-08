package model

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrObjectNameRequired Error = "object name is required"
	ErrObjectNotFound     Error = "object not found"
)
