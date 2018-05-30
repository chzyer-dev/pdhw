package pdhw

import (
	"fmt"
)

// Strategy 是副本放置策略，副本策略用于约束一个 Region 的副本数及副本位置的约束，具体内容由你来设计。以下是一些常见的需求供你参考（这些你可以选择实现一部分）：
//   1. 3 副本随机分布在不同的节点
//   2. 节点分布在多个 Rack 上，3 副本中的任意 2 副本都不能在同一个 Rack
//   3. 3 DC，3 副本分别分布在不同的 DC
//   4. 2 DC，每个 DC 内有多个 Rack，3 副本分布在不同的 Rack 且不能 3 副本都在同一个 DC
//   5. 2 DC，5 副本，其中一个 DC 放置 3 副本，另外一个 DC 放置 2 副本
//   6. 3 副本，要求存储在 ssd 磁盘
//   7. 3 副本至少有一个副本存储在 ssd 磁盘
// 以上策略进行组合
type Strategy struct {
	MaxReplicas int // 总副本数

	// 可选参数, 如果为 0 表示不限制
	MaxDatacenterCnt       int // 最多分布的 Datacenter 数
	MaxReplicaInDatacenter int // 单个 Datacenter 的副本上限
	MaxReplicaInRack       int // 单个 Rack 的副本上限
	MinReplicaRequireSSD   int // 需要放到 SSD 的副本数量
}

func ifOrMax(metric int, val int) int {
	if metric > 0 {
		return val
	}
	return 999
}

const (
	ScorePosMaxReplicaInRack = iota
	ScorePosMaxReplicaInDatacenter
	ScorePosMinReplicaRequireSSD
	ScorePosMaxDatacenterCnt
)

func (s *Strategy) GetScores(info *StoreStrategyInfo) []int {
	return []int{
		ifOrMax(s.MaxReplicaInRack, s.MaxReplicaInRack-info.MaxReplicaInRack),
		ifOrMax(s.MaxReplicaInDatacenter, s.MaxReplicaInDatacenter-info.MaxReplicaInDatacenter),
		ifOrMax(s.MinReplicaRequireSSD, info.ReplicaUsingSSD-s.MinReplicaRequireSSD),
		ifOrMax(s.MaxDatacenterCnt, s.MaxDatacenterCnt-info.DatacenterCnt),
	}
}

func findFirstNegative(scores []int) (idx, score int) {
	for i, s := range scores {
		if s < 0 {
			return i, s
		}
	}
	return len(scores), 0
}

// 尝试找到更优解
func (s *Strategy) CheckSwap(regionStore *RegionStore) (swapOut, swapIn int, ok bool) {
	info := CalculateStoreInfo(regionStore, NilFilter)
	scores := s.GetScores(&info)

	// 找到第一个不符合条件的指标
	minScorePos, minScoreVal := findFirstNegative(scores)
	if minScoreVal >= 0 {
		return 0, 0, false
	}

	excludeFilter := NewExcludeFilter(regionStore.Region.Replicas...)

	switch minScorePos {
	case ScorePosMaxDatacenterCnt:
		// 需要合并数据中心
		// 找到节点数最少的两组数据中心(因为节点数可能一样, 所以是两组而不是两个)
		var cnts []int
		var mins [2]int
		for _, dc := range info.Datacenters {
			cnt := regionStore.Count(NewKVFilter(LabelDatacenter, dc))
			if mins[0] == 0 || cnt < mins[0] {
				mins[1] = mins[0]
				mins[0] = cnt
			} else if mins[1] == 0 || cnt < mins[1] {
				mins[1] = cnt
			}
			cnts = append(cnts, cnt)
		}
		var fromDC, toDC []string
		for idx, dc := range info.Datacenters {
			if cnts[idx] == mins[0] {
				fromDC = append(fromDC, dc)
			} else if cnts[idx] == mins[1] {
				toDC = append(toDC, dc)
			}
		}
		// 确保同样最小数量的DC内部也可以相互迁移
		toDC = append(toDC, fromDC...)
		fromList := regionStore.Filter(NewKVsFilter(LabelDatacenter, fromDC))
		toList := regionStore.FilterAll(Filters{
			NewKVsFilter(LabelDatacenter, toDC), NewExcludeFilter(regionStore.Region.Replicas...),
		})
		return trySwapStoreList(s, minScorePos, minScoreVal, regionStore, fromList, toList)
	case ScorePosMaxReplicaInDatacenter:
		// 去要迁移去其他dc
		// 可能造成没有足够的
		// 从这里面选择一个迁移
		// 首先考虑Rack是否符合
		fromList := regionStore.Filter(NewKVFilter(LabelDatacenter, info.MaxReplicaInDatacenterInfo))
		toList := regionStore.FilterAll(Filters{
			NewNotFilters(NewKVFilter(LabelDatacenter, info.MaxReplicaInDatacenterInfo)),
			excludeFilter,
		})
		return trySwapStoreList(s, minScorePos, minScoreVal, regionStore, fromList, toList)
	case ScorePosMaxReplicaInRack:
		// 迁移出机架
		filters := Filters{
			NewKVFilter(LabelDatacenter, info.MaxReplicaInRackInfo[0]),
			NewKVFilter(LabelRack, info.MaxReplicaInRackInfo[1]),
		}
		fromList := regionStore.Filter(filters)
		toList := regionStore.FilterAll(Filters{
			NewNotFilters(filters), excludeFilter,
		})
		return trySwapStoreList(s, minScorePos, minScoreVal, regionStore, fromList, toList)
	case ScorePosMinReplicaRequireSSD:
		// 找到对应的SSD
		fromList := regionStore.Filter(NewKVFilter(LabelStorageType, ValueHDD))
		toList := regionStore.FilterAll(Filters{
			excludeFilter, NewKVFilter(LabelStorageType, ValueSSD)})
		return trySwapStoreList(s, minScorePos, minScoreVal, regionStore, fromList, toList)
	}

	return 0, 0, false
}

type StoreStrategyInfo struct {
	DatacenterCnt              int // 分布的 Datacenter 数
	MaxReplicaInDatacenter     int // 最大的单个 Datacenter 的副本数
	MaxReplicaInDatacenterInfo string
	MaxReplicaInRack           int // 最大的单个 Rack 的副本数
	MaxReplicaInRackInfo       []string
	ReplicaUsingSSD            int // 使用SSD的副本数

	Datacenters []string
}

func (s StoreStrategyInfo) String() string {
	return fmt.Sprintf("%#v", &s)
}
