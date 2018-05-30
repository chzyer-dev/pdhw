package pdhw

import "testing"

func TestFilter(t *testing.T) {
	filter := NewExcludeFilter(1, 2)
	if !filter.Filter(&Store{ID: 1}) {
		t.Fatal("!filter.Filter(&Store{ID: 1})")
	}
	if !filter.Filter(&Store{ID: 2}) {
		t.Fatal("!filter.Filter(&Store{ID: 2})")
	}
	if filter.Filter(&Store{ID: 3}) {
		t.Fatal("filter.Filter(&Store{ID: 3})")
	}

	filter = NewKVFilter("k1", "v1")
	if filter.Filter(&Store{Labels: map[string]string{"k1": "v1"}}) {
		t.Fatal(`filter.Filter(&Store{Labels: map[string]string{"k1": "v1"}})`)
	}
	if !filter.Filter(&Store{Labels: map[string]string{"k1": "v2"}}) {
		t.Fatal(`!filter.Filter(&Store{Labels: map[string]string{"k1": "v2"}})`)
	}
}
