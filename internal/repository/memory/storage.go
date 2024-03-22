package memory

import (
	"context"
	"sync"

	"github.com/Longreader/go-shortener-url.git/internal/repository"
	"github.com/Longreader/go-shortener-url.git/internal/tools"
	"github.com/sirupsen/logrus"
)

type MemoryStorage struct {
	sync.RWMutex
	LinksStorage    map[repository.URL]repository.ID
	UserLinkStorage map[repository.ID]repository.LinkData
}

// Constructor for MemoryStorage type
func NewMemoryStorage() (*MemoryStorage, error) {
	st := &MemoryStorage{
		LinksStorage:    make(map[repository.URL]repository.ID),
		UserLinkStorage: make(map[repository.ID]repository.LinkData),
	}
	return st, nil
}

// Set method for MemoryStorage storage
func (st *MemoryStorage) Set(
	_ context.Context,
	url repository.URL,
	user repository.User,
) (id repository.ID, err error) {
	return st.SetLink(url, user)
}

// Set method for Storage
func (st *MemoryStorage) SetLink(
	url repository.URL,
	user repository.User,
) (repository.ID, error) {

	st.Lock()
	defer st.Unlock()

	value, ok := st.LinksStorage[url]
	if ok {
		return value, repository.ErrURLAlreadyExists
	}

	id, err := tools.RandStringBytes(5)
	if err != nil {
		return "", err
	}

	for exists := true; exists; _, exists = st.UserLinkStorage[id] {
		id, err = tools.RandStringBytes(5)
		if err != nil {
			return "", err
		}
	}

	st.UserLinkStorage[id] = repository.LinkData{
		URL:  url,
		User: user,
	}
	st.LinksStorage[url] = id

	return id, nil
}

// Get method fot MemoryStorage storage
func (st *MemoryStorage) Get(_ context.Context, id repository.ID) (url repository.URL, deleted bool, err error) {

	st.RLock()
	defer st.RUnlock()

	data, ok := st.UserLinkStorage[id]
	if ok {
		return data.URL, data.Deleted, nil
	}

	return "", false, repository.ErrURLNotFound
}

func (st *MemoryStorage) Delete(_ context.Context, ids []repository.ID, user repository.User) error {

	st.RWMutex.Lock()
	defer st.RWMutex.Unlock()
	for _, id := range ids {
		link, ok := st.UserLinkStorage[id]
		if !ok {
			logrus.Error("user does not exist")
			continue
		}
		if link.User != user {
			logrus.Error("error wrong user")
			continue
		}
		// link.Deleted = true
		// st.UserLinkStorage[id] = link
		st.DeleteLink(link, id)
	}
	return nil
}

func (st *MemoryStorage) DeleteLink(link repository.LinkData, id repository.ID) {
	link.Deleted = true
	st.UserLinkStorage[id] = link
}

func (st *MemoryStorage) GetAllByUser(_ context.Context, user repository.User) (data []repository.LinkData, err error) {

	st.RLock()
	defer st.RUnlock()

	data = make([]repository.LinkData, 0)

	for id, value := range st.UserLinkStorage {
		if value.User != user {
			continue
		}
		if value.Deleted {
			continue
		}

		data = append(data, repository.LinkData{
			ID:   id,
			URL:  value.URL,
			User: value.User,
		})
	}

	return data, nil
}

func (st *MemoryStorage) RunDelete() {
	//methon implement interface
}

// Check connection with Storage
func (st *MemoryStorage) Ping(_ context.Context) (bool, error) {
	return true, nil
}

func (st *MemoryStorage) Close(_ context.Context) error {
	return nil
}
