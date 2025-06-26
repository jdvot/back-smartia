package models

import (
	"sync"
	"time"
)

type Document struct {
	ID            string
	UserID        string
	Filename      string
	Size          int
	MimeType      string
	UploadDate    time.Time
	Status        string
	OcrText       *string
	Summary       *string
	OcrStatus     string
	SummaryStatus string
}

type DocumentStore struct {
	mu    sync.RWMutex
	docs  map[string]*Document
}

var store = &DocumentStore{
	docs: make(map[string]*Document),
}

func GetStore() *DocumentStore {
	return store
}

func (s *DocumentStore) Add(doc *Document) {
	s.mu.Lock()
	s.docs[doc.ID] = doc
	s.mu.Unlock()
}

func (s *DocumentStore) Get(id string) (*Document, bool) {
	s.mu.RLock()
	d, ok := s.docs[id]
	s.mu.RUnlock()
	return d, ok
}

func (s *DocumentStore) Delete(id string) {
	s.mu.Lock()
	delete(s.docs, id)
	s.mu.Unlock()
}

func (s *DocumentStore) ListByUser(userID string) []*Document {
	s.mu.RLock()
	var result []*Document
	for _, d := range s.docs {
		if d.UserID == userID {
			result = append(result, d)
		}
	}
	s.mu.RUnlock()
	return result
} 