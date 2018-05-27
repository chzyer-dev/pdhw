package pdhw

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	stores := []Store{
		{1, StoreLabelInit("dc1", "rack1", "d1r1h1", ValueHDD)}, //
		{2, StoreLabelInit("dc2", "rack1", "d2r1h1", ValueHDD)},
		{3, StoreLabelInit("dc3", "rack1", "d3r1h1", ValueHDD)}, //
		{4, StoreLabelInit("dc3", "rack1", "d3r1h1", ValueSSD)},
		{5, StoreLabelInit("dc3", "rack1", "d3r1h2", ValueSSD)},
		{6, StoreLabelInit("dc3", "rack2", "d3r2h3", ValueSSD)}, //
	}
	strategy := Strategy{
		MaxReplicas:            3,
		MaxDatacenterCnt:       2,
		MaxRackCnt:             0,
		MaxReplicaInDatacenter: 2,
		MaxReplicaInRack:       1,
		MinReplicaRequireSSD:   1,
	}
	region := Region{[]int{2, 1, 4}}
	regionStores := NewRegionStore(stores, region)
	storeInfo := strategy.CalculateStoreInfo(regionStores, NilFilter)
	fmt.Println(storeInfo)
	score := storeInfo.CalculateScore(&strategy)
	fmt.Println(score.Score(), score)
	newRegion := Check(stores, region, strategy)
	// stores 可能失效了
	newInfo := strategy.CalculateStoreInfo(NewRegionStore(stores, newRegion), NilFilter)
	fmt.Println("final:", newRegion, "origin:", region)
	fmt.Println(newInfo)
	fmt.Println(newInfo.CalculateScore(&strategy))
}

func TestDecay(t *testing.T) {
	fmt.Println(decayMax(1, 3, 20))
}
