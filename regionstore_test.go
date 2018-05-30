package pdhw

import (
	"reflect"
	"testing"
)

func TestRegionStoreMapValue(t *testing.T) {
	regionStore := NewRegionStore(Stores{
		Store{1, map[string]string{"k": "v1"}},
		Store{2, map[string]string{"k": "v1"}},
		Store{3, map[string]string{"k": "v2"}},
		Store{4, map[string]string{"k": "v3"}},
	}, Region{[]int{1, 2, 3}})
	val := regionStore.MapValue("k", NewExcludeFilter(3))
	if !reflect.DeepEqual(val, []string{"v1"}) {
		t.Fatal(`!reflect.DeepEqual(val, []string{"v1"})`)
	}
	val = regionStore.MapValue("k", NilFilter)
	if !reflect.DeepEqual(val, []string{"v1", "v2"}) {
		t.Fatal(`!reflect.DeepEqual(val, []string{"v1", "v2"})`)
	}
	regionStore.Tmp = &regionStore.Stores[3]
	val = regionStore.MapValue("k", NilFilter)
	if !reflect.DeepEqual(val, []string{"v1", "v2", "v3"}) {
		t.Fatal(`!reflect.DeepEqual(val, []string{"v1", "v2", "v3"})`)
	}
}
