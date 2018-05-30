package pdhw

import "fmt"

type Stores []Store

func (s Stores) Index(id int) int {
	for idx := range s {
		if s[idx].ID == id {
			return idx
		}
	}
	return -1
}

type Store struct {
	ID     int
	Labels map[string]string
}

func (s *Store) String() string {
	return fmt.Sprintf("{%v}", s.ID)
}

func (s *Store) IsValid() bool {
	return s.ID > 0
}

func (s *Store) SetLabel(key string, value string) {
	if s.Labels == nil {
		s.Labels = make(map[string]string, 8)
	}
	s.Labels[key] = value
}

func (s *Store) GetLabel(key string) string {
	return s.Labels[key]
}
