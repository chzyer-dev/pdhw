package pdhw

type Region struct {
	Replicas []int
}

// 获取副本的节点信息
func (r Region) GetStores(stores Stores) []Store {
	ret := make([]Store, len(r.Replicas))
	for idx, id := range r.Replicas {
		storeIdx := stores.Index(id)
		if storeIdx >= 0 {
			ret[idx] = stores[storeIdx]
		}
	}
	return ret
}

func (r *Region) Swap(fromID int, toID int) bool {
	for idx, replica := range r.Replicas {
		if replica == fromID {
			r.Replicas[idx] = toID
			return true
		}
	}
	return false
}

func (r Region) Clone() Region {
	replicas := make([]int, len(r.Replicas))
	copy(replicas, r.Replicas)
	return Region{
		Replicas: replicas,
	}
}
