package banners

import (
	"context"
	"errors"
	"sync"
)

type Service struct {
	mu    sync.RWMutex
	items []*Banner
	MaxID int64
}

type Banner struct {
	ID      int64
	Title   string
	Content string
	Button  string
	Link    string
	Image   string
}

func NewService() *Service {
	return &Service{items: make([]*Banner, 0), MaxID: 1}
}

func (s *Service) All(ctx context.Context) ([]*Banner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.items, nil
}

func (s *Service) ByID(ctx context.Context, id int64) (*Banner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, banner := range s.items {
		if banner.ID == id {
			return banner, nil
		}
	}

	return nil, errors.New("item not found")
}

func (s *Service) Save(ctx context.Context, item *Banner) (*Banner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = append(s.items, item)

	return s.items[len(s.items)-1], nil
}

func (s *Service) RemoveByID(ctx context.Context, id int64) (*Banner, error) {

	item, err := s.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ind := 0
	for _, banner := range s.items {
		if banner.ID == id {
			break
		}
		ind++
	}
	s.items = removeIndex(s.items, ind)

	return item, nil
}

func removeIndex(s []*Banner, index int) []*Banner {
	return append(s[:index], s[index+1:]...)
}
