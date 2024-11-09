package entity

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/ssimpl/simple-storage/internal/api/model"
)

type Server struct {
	bun.BaseModel `bun:"table:servers"`

	ID        uuid.UUID `bun:"id,pk,nullzero"`
	Addr      string    `bun:"addr"`
	UsedSpace int64     `bun:"used_space"`
}

func (s Server) ToModel() model.Server {
	return model.Server{
		ID:        s.ID,
		Addr:      s.Addr,
		UsedSpace: s.UsedSpace,
	}
}

func ServerFromModel(m model.Server) Server {
	return Server{
		ID:        m.ID,
		Addr:      m.Addr,
		UsedSpace: m.UsedSpace,
	}
}
