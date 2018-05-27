package pdhw

import (
	"fmt"
	"math"
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
	MaxRackCnt             int // 最多分布的 Rack 数
	MaxReplicaInDatacenter int // 单个 Datacenter 的副本上限
	MaxReplicaInRack       int // 单个 Rack 的副本上限
	MinReplicaRequireSSD   int // 需要放到 SSD 的副本数量
}

type StoreStrategyInfo struct {
	DatacenterCnt          int // 分布的 Datacenter 数
	RackCnt                int // 分布的 Rack 数
	MaxReplicaInDatacenter int // 最大的单个 Datacenter 的副本数
	MaxReplicaInRack       int // 最大的单个 Rack 的副本数
	ReplicaUsingSSD        int // 使用SSD的副本数
}

func (info *StoreStrategyInfo) CalculateScore(s *Strategy) *StrategyScore {
	return &StrategyScore{
		decayMax(info.DatacenterCnt, s.MaxDatacenterCnt, 20),
		decayMax(info.RackCnt, s.MaxRackCnt, 20),
		decayMax(info.MaxReplicaInDatacenter, s.MaxReplicaInDatacenter, 20),
		decayMax(info.MaxReplicaInRack, s.MaxReplicaInRack, 20),
		decayMax2(info.ReplicaUsingSSD, s.MinReplicaRequireSSD, 20),
	}
}

func (s StoreStrategyInfo) String() string {
	return fmt.Sprintf("%#v", &s)
}

type StrategyScore struct {
	MaxDatacenterCntScore       int
	MaxRackCntScore             int
	MaxReplicaInDatacenterScore int
	MaxReplicaInRackScore       int
	MinReplicaRequireSSDScore   int
}

func (s StrategyScore) String() string {
	return fmt.Sprintf("%#v", &s)
}

func (s *StrategyScore) Score() int {
	return s.MaxDatacenterCntScore + s.MaxRackCntScore +
		s.MaxReplicaInDatacenterScore +
		s.MaxReplicaInRackScore +
		s.MinReplicaRequireSSDScore
}

// 检查配置是否有问题
func (s *Strategy) CheckConfig() error {
	// 检查配置本身一些不可能性

	return nil
}

func (s *Strategy) CalculateStoreInfo(regionStore *RegionStore, filter Filter) (info StoreStrategyInfo) {
	dcs := regionStore.MapValue(LabelDatacenter, filter)
	info.DatacenterCnt = len(dcs)

	filters := make(Filters, 1, 3)
	filters[0] = filter
	for _, dc := range dcs {
		dcFilter := append(filters, NewKVFilter(LabelDatacenter, dc))
		racks := regionStore.MapValue(LabelRack, dcFilter)
		info.RackCnt += len(racks)
		info.MaxReplicaInDatacenter = max(info.MaxReplicaInDatacenter, regionStore.Count(dcFilter))
		for _, rack := range racks {
			rackFilter := append(dcFilter, NewKVFilter(LabelRack, rack))
			info.MaxReplicaInRack = max(info.MaxReplicaInRack, regionStore.Count(rackFilter))
		}
	}

	info.ReplicaUsingSSD = regionStore.Count(append(filters, NewKVFilter(LabelStorageType, ValueSSD)))

	return info
}

// 评分标准
// 目标是将对结果影响最大的节点找出来
// 所以, 评分需要按照一下规则设计:
//   需要计算偏离程度. 找出多项指标综合偏差程度大的节点
//   所以, 假定每项分数 20 分, 5项刚好 100 分. 衰减程度按副本数来决定
func (s *Strategy) CalculateBalanceScore(regionStore *RegionStore, filter Filter) int {
	info := s.CalculateStoreInfo(regionStore, filter)
	return info.CalculateScore(s).Score()
}

func (s *Strategy) CalculateBalanceScoreWith(regionStore *RegionStore, filter Filter, store *Store) int {
	regionStore.Tmp = store
	info := s.CalculateStoreInfo(regionStore, filter)
	regionStore.Tmp = nil
	return info.CalculateScore(s).Score()
}

func decayMax2(fieldvalue, origin int, maxScore int) int {
	if origin == 0 {
		return 0
	}
	originF64 := float64(origin)

	score := int(decayF64(float64(fieldvalue), originF64, originF64) * float64(maxScore))
	return score
}

func decayMax(fieldvalue, origin int, maxScore int) int {
	if origin == 0 {
		return 0
	}
	originF64 := float64(origin)

	score := int(decayF64(float64(fieldvalue), originF64, originF64) * float64(maxScore))
	if fieldvalue > origin {
		score -= maxScore
	}
	return score
}

func decayF64(fieldvalue, origin float64, scale float64) float64 {
	return math.Max(0, (scale-math.Max(math.Abs(fieldvalue-origin), 0))/scale)
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *Strategy) findBestStore(regionStore *RegionStore, swapOutID int) (*Store, int) {
	var bestStore *Store
	var bestScore int
	filter := NewExcludeFilter(swapOutID)
	regionFilter := NewExcludeFilter(regionStore.Region.Replicas...)
	for idx := range regionStore.Stores {
		store := &regionStore.Stores[idx]
		if regionFilter.Filter(store) {
			// 过滤掉已经选择的store
			continue
		}
		score := s.CalculateBalanceScoreWith(regionStore, filter, store)
		if score > bestScore || bestStore == nil {
			bestStore = store
			bestScore = score
		}
	}
	return bestStore, bestScore
}

var locationLabels = []string{LabelDatacenter, LabelRack, LabelHost, LabelStorageType}

func CheckAndFixOne(stores []Store, region Region, strategy Strategy) (Region, bool) {
	regionStore := NewRegionStore(stores, region)
	oldScore := strategy.CalculateBalanceScore(regionStore, NilFilter)

	var (
		maxScore int = oldScore
		swapCost int
		swapIn   *Store
		swapOut  *Store
	)

	for _, idx := range regionStore.Indices {
		store := &regionStore.Stores[idx]
		newStore, newScore := strategy.findBestStore(regionStore, store.ID)
		fmt.Printf("swap: [%v -> %v], score: %v, cost: %v\n", store.ID, newStore.ID, newScore,
			calculateSwapCost(locationLabels, store.ID, newStore.ID, stores))

		if newScore > maxScore {
			cost := calculateSwapCost(locationLabels, store.ID, newStore.ID, stores)
			fmt.Printf("swap: [%v -> %v], score: %v, cost: %v\n", store.ID, newStore.ID, newScore, cost)
			if cost < swapCost || swapOut == nil {
				maxScore = newScore
				swapCost = cost
				swapOut = store
				swapIn = newStore
			}
		}
	}

	if swapOut != nil {
		newRegion := region.Clone()
		newRegion.Swap(swapOut.ID, swapIn.ID)
		println(fmt.Sprintf("swap [%v -> %v]", swapOut.ID, swapIn.ID))
		return newRegion, true
	}

	return region, false
}

func calculateSwapCost(labels []string, oldID, newID int, stores Stores) int {
	oldStore := stores.Get(oldID)
	newStore := stores.Get(newID)
	for idx, label := range labels {
		if newStore.GetLabel(label) != oldStore.GetLabel(label) {
			return len(labels) - idx
		}
	}
	return 0
}

// Check 函数用于检查一个 Region 是否满足策略的约束，
// 如果不满足则需要返回新的副本分布使其尽可能满足约束
// （Region 可能不满足策略约束的原因可能是策略调整或者节点有变动）。
// 注意应尽量减少新 Region 与原 Region 的差异来减少调度开销。
// 参数中 stores 是集群中的所有存储节点，region 标识了当前的副本分布情况，
// strategy 是这个 Region 对应的 placement 策略。
func Check(stores []Store, region Region, strategy Strategy) Region {
	// 当前的节点信息
	ok := true
	for i := 0; i < len(region.Replicas) && ok; i++ {
		region, ok = CheckAndFixOne(stores, region, strategy)
	}
	if ok {
		// 相当于大量的调度操作, 需要记录
	}
	return region
}
