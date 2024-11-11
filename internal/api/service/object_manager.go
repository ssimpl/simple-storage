package service

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/google/uuid"

	"github.com/ssimpl/simple-storage/internal/api/model"
)

type objectStorage interface {
	Store(ctx context.Context, serverAddr string, objectID uuid.UUID, data io.Reader) error
	Retrieve(ctx context.Context, serverAddr string, objectID uuid.UUID, dst io.Writer) error
}

type metaRepository interface {
	SaveObjectMeta(ctx context.Context, meta model.ObjectMeta) error
	GetObjectMeta(ctx context.Context, objectName string) (model.ObjectMeta, error)
	GetServers(ctx context.Context) ([]model.Server, error)
}

type ObjectManager struct {
	objectStorage objectStorage
	metaRepo      metaRepository
	fragmentCount int
}

func NewObjectManager(
	objectStorage objectStorage, metaRepo metaRepository, fragmentCount int,
) *ObjectManager {
	return &ObjectManager{
		objectStorage: objectStorage,
		metaRepo:      metaRepo,
		fragmentCount: fragmentCount,
	}
}

// TODO: use the same servers for fragments if object with specified name already exists
func (m *ObjectManager) StoreObject(
	ctx context.Context, objectName string, src io.Reader, size int64,
) error {
	servers, err := m.metaRepo.GetServers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get servers: %w", err)
	}
	if len(servers) == 0 {
		return fmt.Errorf("no servers available")
	}

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].UsedSpace < servers[j].UsedSpace
	})

	fragmentSize := size / int64(m.fragmentCount)
	lastFragmentSize := size - (fragmentSize * int64(m.fragmentCount-1))

	metaFragments := make([]model.ObjectFragmentMeta, 0, m.fragmentCount)
	for i := 0; i < m.fragmentCount; i++ {
		fragmentIndex := i

		server := servers[fragmentIndex%len(servers)]

		currentFragmentSize := fragmentSize
		if fragmentIndex == m.fragmentCount-1 {
			currentFragmentSize = lastFragmentSize
		}

		fragmentID := getFragmentID(objectName, fragmentIndex)
		fragmentReader := io.LimitReader(src, currentFragmentSize)

		//TODO: implement retries
		err := m.objectStorage.Store(ctx, server.Addr, fragmentID, fragmentReader)
		if err != nil {
			return fmt.Errorf("failed to store fragment: %w", err)
		}

		metaFragments = append(metaFragments, model.ObjectFragmentMeta{
			SeqNum:       fragmentIndex,
			ServerID:     server.ID,
			FragmentID:   fragmentID,
			FragmentSize: currentFragmentSize,
		})
	}

	return m.metaRepo.SaveObjectMeta(ctx, model.ObjectMeta{
		ObjectName: objectName,
		Fragments:  metaFragments,
	})
}

func getFragmentID(objectName string, seqNum int) uuid.UUID {
	return uuid.NewSHA1(uuid.UUID{}, []byte(fmt.Sprintf("%s-%d", objectName, seqNum)))
}

func (m *ObjectManager) RetrieveObject(ctx context.Context, objectName string, dst io.Writer) error {
	meta, err := m.metaRepo.GetObjectMeta(ctx, objectName)
	if err != nil {
		return fmt.Errorf("failed to get object meta: %w", err)
	}

	sort.Slice(meta.Fragments, func(i, j int) bool {
		return meta.Fragments[i].SeqNum < meta.Fragments[j].SeqNum
	})

	servers, err := m.metaRepo.GetServers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get servers: %w", err)
	}
	if len(servers) == 0 {
		return fmt.Errorf("no servers available")
	}

	serversByID := make(map[uuid.UUID]model.Server, len(servers))
	for _, s := range servers {
		serversByID[s.ID] = s
	}

	for _, f := range meta.Fragments {
		server, ok := serversByID[f.ServerID]
		if !ok {
			return fmt.Errorf("server '%s' not found: %w", f.ServerID, model.ErrServerNotFound)
		}

		//TODO: implement retries
		if err := m.objectStorage.Retrieve(ctx, server.Addr, f.FragmentID, dst); err != nil {
			return fmt.Errorf("failed to retrieve fragment '%s': %w", f.FragmentID, err)
		}
	}

	return nil
}
