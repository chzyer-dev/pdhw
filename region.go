package pdhw

type Region struct {
	Replicas []int
}

func (r Region) Clone() Region {
	replicas := make([]int, len(r.Replicas))
	copy(replicas, r.Replicas)
	return Region{
		Replicas: replicas,
	}
}
