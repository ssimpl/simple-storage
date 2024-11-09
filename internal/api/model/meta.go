package model

import "github.com/google/uuid"

type ObjectMeta struct {
	ObjectName string
	Fragments  []ObjectFragmentMeta
}

type ObjectFragmentMeta struct {
	SeqNum       int
	ServerID     uuid.UUID
	FragmentID   uuid.UUID
	FragmentSize int64
}
