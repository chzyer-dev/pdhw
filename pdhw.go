package pdhw

func Check(stores []Store, region Region, strategy Strategy) Region {
	rs := NewRegionStore(stores, region)
	for {
		swapOut, swapIn, ok := strategy.CheckSwap(rs)
		if !ok {
			break
		}
		rs.Swap(swapOut, swapIn)
	}

	return rs.Region
}

// 按照 Strategy 的维度, 算出当前副本集的每个维度的指标信息
func CalculateStoreInfo(regionStore *RegionStore, filter Filter) (info StoreStrategyInfo) {
	dcs := regionStore.MapValue(LabelDatacenter, filter)
	info.Datacenters = dcs
	info.DatacenterCnt = len(dcs)

	filters := make(Filters, 1, 3)
	filters[0] = filter
	for _, dc := range dcs {
		dcFilter := append(filters, NewKVFilter(LabelDatacenter, dc))
		racks := regionStore.MapValue(LabelRack, dcFilter)
		cnt := regionStore.Count(dcFilter)
		if cnt > info.MaxReplicaInDatacenter {
			info.MaxReplicaInDatacenter = cnt
			info.MaxReplicaInDatacenterInfo = dc
		}
		for _, rack := range racks {
			rackFilter := append(dcFilter, NewKVFilter(LabelRack, rack))
			cnt := regionStore.Count(rackFilter)
			if cnt > info.MaxReplicaInRack {
				info.MaxReplicaInRack = cnt
				info.MaxReplicaInRackInfo = []string{dc, rack}
			}
		}
	}

	info.ReplicaUsingSSD = regionStore.Count(append(filters, NewKVFilter(LabelStorageType, ValueSSD)))

	return info
}

var locationLabels = []string{LabelDatacenter, LabelRack, LabelHost}

// 计算一次数据迁移的代价
func calculateSwapCost(labels []string, from, to *Store) int {
	if from.ID == to.ID {
		return 0
	}
	for idx, label := range labels {
		if to.GetLabel(label) != from.GetLabel(label) {
			return (len(labels) - idx) + 1
		}
	}
	return 1
}

func trySwapStoreList(s *Strategy, minScorePos, minScoreVal int, regionStore *RegionStore, fromList, toList []*Store) (from, to int, ok bool) {
	minCost := -1
	for _, fromStore := range fromList {
		for _, toStore := range toList {
			if fromStore.ID == toStore.ID {
				continue
			}
			regionStore.Tmp = toStore
			info := CalculateStoreInfo(regionStore, NewExcludeFilter(fromStore.ID))
			regionStore.Tmp = nil
			cost := calculateSwapCost(locationLabels, fromStore, toStore)
			scores := s.GetScores(&info)
			idx, score := findFirstNegative(scores)
			// fmt.Println(fromStore.ID, toStore.ID, ":", idx, score, scores)

			isScoreBetter := (score > minScoreVal) && (minScoreVal < 0)
			isScoreEqual := (score == minScoreVal) || (minScoreVal > 0 && score > 0)
			isIdxBetter := (idx > minScorePos)
			isIdxEqual := idx == minScorePos
			isCostBetter := (minCost == -1) || cost < minCost

			// 如果分数变得更好, 或者能将改问题转换成下个问题, 就直接替换
			// 在情况没有变得更好的情况下, 只有cost更低才会采用
			// 当然如果情况变得更差, 就不考虑了
			if isScoreBetter || (isScoreEqual && isIdxBetter) || (isScoreEqual && isIdxEqual && isCostBetter) {
				minScorePos = idx
				minCost = cost
				minScoreVal = score
				from = fromStore.ID
				to = toStore.ID
				ok = true
			}
		}
	}
	return
}
