package crawler

import "sync"

type DetailsStorage struct {
	sync.Mutex
	store map[string]*AlbumDetails
}

func NewDetailsStorage() *DetailsStorage {
	return &DetailsStorage{store: make(map[string]*AlbumDetails)}
}

func (s *DetailsStorage) DetailsForUrl(url string) *AlbumDetails {
	s.Lock()
	defer s.Unlock()
	return s.store[url]
}

func (s *DetailsStorage) Add(details *AlbumDetails) {
	s.Lock()
	s.store[details.coverUrl] = details
	s.Unlock()
}

func (s *DetailsStorage) Remove(details *AlbumDetails) {
	s.Lock()
	delete(s.store, details.coverUrl)
	s.Unlock()
}
