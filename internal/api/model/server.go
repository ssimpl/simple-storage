package model

import "github.com/google/uuid"

type Server struct {
	ID   uuid.UUID
	Addr string
}
