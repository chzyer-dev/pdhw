package pdhw

type RegionStore struct {
	Stores Stores
	Region Region
	// 储存了每个 Region 里面对应 Store 在 Stores 里面的下标
	Index []int

	// 临时加入遍历列表, 可以计算加入新节点后的数量
	Tmp *Store
}

func NewRegionStore(stores Stores, region Region) *RegionStore {
	rs := &RegionStore{
		Stores: stores,
		Region: region.Clone(),
		Index:  make([]int, len(region.Replicas), len(region.Replicas)+1),
	}
	for idx, replica := range region.Replicas {
		rs.Index[idx] = stores.Index(replica)
	}
	return rs
}

func (r *RegionStore) Swap(oldID, newID int) {
	for idx, id := range r.Region.Replicas {
		if id == oldID {
			r.Region.Replicas[idx] = newID
			r.Index[idx] = r.Stores.Index(newID)
		}
	}
}

// 计算副本集里面符合条件的 Store 数量
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

func (r *RegionStore) FilterAll(filter Filter) []*Store {
	vals := make([]*Store, 0, len(r.Stores))
	for i := range r.Stores {
		if filter.Filter(&r.Stores[i]) {
			continue
		}
		vals = append(vals, &r.Stores[i])
	}
	return vals
}

func (r *RegionStore) Filter(filter Filter) []*Store {
	vals := make([]*Store, 0, len(r.Index))
	for _, idx := range r.getAll() {
		store := &r.Stores[idx]
		if filter.Filter(store) {
			continue
		}
		vals = append(vals, store)
	}
	return vals
}

// 根据 Filter 遍历副本集里面 Stores, 取出他们对应 key 的值
// 一般用于计算对应副本集里面对应 datacenter 和 rack 的值
func (r *RegionStore) MapValue(key string, filter Filter) []string {
	vals := make([]string, 0, len(r.Index))
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

// 抽象掉副本集列表, 可以额外插入一个 Store
func (r *RegionStore) getAll() []int {
	if r.Tmp == nil {
		return r.Index
	}
	idx := r.Stores.Index(r.Tmp.ID)
	return append(r.Index, idx)
}
