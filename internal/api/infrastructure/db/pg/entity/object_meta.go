package entity

import (
	"encoding/json"
	"fmt"

	"github.com/ssimpl/simple-storage/internal/api/model"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ObjectMeta struct {
	bun.BaseModel `bun:"table:objects_metadata"`

	Name      string          `bun:"name,pk"`
	Fragments json.RawMessage `bun:"fragments"`
}

type objectMetaFragment struct {
	SeqNum     int       `json:"seq_num"`
	ServerID   uuid.UUID `json:"server_id"`
	FragmentID uuid.UUID `json:"fragment_id"`
}

func (m ObjectMeta) ToModel() (model.ObjectMeta, error) {
	var fragments []objectMetaFragment
	err := json.Unmarshal(m.Fragments, &fragments)
	if err != nil {
		return model.ObjectMeta{}, fmt.Errorf("unmarshal fragments: %w", err)
	}

	modelFragments := make([]model.ObjectFragmentMeta, 0, len(fragments))
	for _, f := range fragments {
		modelFragments = append(modelFragments, model.ObjectFragmentMeta{
			SeqNum:     f.SeqNum,
			ServerID:   f.ServerID,
			FragmentID: f.FragmentID,
		})
	}

	return model.ObjectMeta{
		ObjectName: m.Name,
		Fragments:  modelFragments,
	}, nil
}

func ObjectMetaFromModel(m model.ObjectMeta) (ObjectMeta, error) {
	fragments := make([]objectMetaFragment, 0, len(m.Fragments))
	for _, f := range m.Fragments {
		fragments = append(fragments, objectMetaFragment{
			SeqNum:     f.SeqNum,
			ServerID:   f.ServerID,
			FragmentID: f.FragmentID,
		})
	}

	fragmentsData, err := json.Marshal(fragments)
	if err != nil {
		return ObjectMeta{}, fmt.Errorf("marshal fragments: %w", err)
	}

	return ObjectMeta{
		Name:      m.ObjectName,
		Fragments: fragmentsData,
	}, nil
}
