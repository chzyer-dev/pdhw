package pdhw

var NilFilter Filters

type Filters []Filter

func (fs Filters) Filter(s *Store) bool {
	for _, f := range fs {
		if f.Filter(s) {
			return true
		}
	}
	return false
}

type Filter interface {
	Filter(s *Store) bool
}

func NewExcludeFilter(storeID ...int) *SimpleFilter {
	return &SimpleFilter{ExcludeID: storeID}
}

func NewKVFilter(key, value string) Filter {
	return &SimpleFilter{Key: key, Value: value}
}

type SimpleFilter struct {
	ExcludeID []int
	Key       string
	Value     string
}

func (f *SimpleFilter) Filter(s *Store) bool {
	if len(f.ExcludeID) > 0 {
		for _, id := range f.ExcludeID {
			if s.ID == id {
				return true
			}
		}
	}
	if f.Key != "" {
		return s.GetLabel(f.Key) != f.Value
	}
	return false
}
