package pdhw

type RegionStore struct {
	Stores  Stores
	Region  Region
	Indices []int

	Tmp *Store
}

func NewRegionStore(stores Stores, region Region) *RegionStore {
	rs := &RegionStore{
		Stores:  stores,
		Region:  region.Clone(),
		Indices: make([]int, len(region.Replicas)),
	}
	for idx, replica := range region.Replicas {
		rs.Indices[idx] = stores.Index(replica)
	}
	return rs
}

func (r *RegionStore) GetStores(filter Filter) (ids []int) {
	for _, idx := range r.getAll() {
		store := &r.Stores[idx]
		if filter.Filter(store) {
			continue
		}
		ids = append(ids, store.ID)
	}
	return ids
}

func (r *RegionStore) Swap(oldID, newID int) {
	for idx, id := range r.Region.Replicas {
		if id == oldID {
			r.Region.Replicas[idx] = newID
			r.Indices[idx] = r.Stores.Index(newID)
		}
	}
}

func (r *RegionStore) MapValue(key string, filter Filter) []string {
	vals := make([]string, 0, len(r.Indices))
loop:
	for _, idx := range r.getAll() {
		store := &r.Stores[idx]
		if filter.Filter(store) {
			continue
		}

		storeValue := store.GetLabel(key)
		for _, val := range vals {
			if val == storeValue {
				continue loop
			}
		}
		vals = append(vals, storeValue)
	}
	return vals
}

func (r *RegionStore) Get(idx int) *Store {
	return &r.Stores[r.Indices[idx]]
}

func (r *RegionStore) getAll() []int {
	if r.Tmp == nil {
		return r.Indices
	}
	idx := r.Stores.Index(r.Tmp.ID)
	return append(r.Indices, idx)
}

func (r *RegionStore) Count(filter Filter) int {
	var val int
	for _, idx := range r.getAll() {
		store := &r.Stores[idx]
		if filter.Filter(store) {
			continue
		}
		val++
	}
	return val
}
