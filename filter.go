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

type NotFilters []Filter

func (fs NotFilters) Filter(s *Store) bool {
	return !Filters(fs).Filter(s)
}

func NewNotFilters(fs ...Filter) NotFilters {
	return NotFilters(fs)
}

func NewExcludeFilter(storeID ...int) Filter {
	return &SimpleFilter{ExcludeID: storeID}
}

func NewKVFilter(key, value string) Filter {
	return &SimpleFilter{Key: key, Value: value}
}

func NewKVsFilter(key string, values []string) Filter {
	return &SimpleFilter{Key: key, Values: values}
}

type SimpleFilter struct {
	ExcludeID []int
	Key       string
	Value     string
	Values    []string
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
		if len(f.Values) > 0 {
			val := s.GetLabel(f.Key)
			for _, n := range f.Values {
				if n == val {
					return false
				}
			}
			return true
		}

		return s.GetLabel(f.Key) != f.Value
	}
	return false
}
