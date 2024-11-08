package model

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrObjectNotFound   Error = "object not found"
	ErrDBMalfunctioning Error = "db malfunctioning"
	ErrServerNotFound   Error = "server not found"
)
