package pdhw

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestStrategyInfo(t *testing.T) {
	stores := []Store{
		{1, StoreLabelInit("dc1", "rack1", "d1r1h1", ValueHDD)}, //
		{2, StoreLabelInit("dc2", "rack1", "d2r1h1", ValueHDD)},
		{3, StoreLabelInit("dc3", "rack1", "d3r1h1", ValueHDD)}, //
		{4, StoreLabelInit("dc3", "rack1", "d3r1h1", ValueSSD)},
		{5, StoreLabelInit("dc3", "rack1", "d3r1h2", ValueSSD)},
		{6, StoreLabelInit("dc3", "rack2", "d3r2h3", ValueSSD)}, //
	}
	info := CalculateStoreInfo(NewRegionStore(stores, Region{
		Replicas: []int{1, 2, 3, 4, 5, 6},
	}), NilFilter)
	if info.DatacenterCnt != 3 {
		t.Fatal("dccnt != 3")
	}
	if info.MaxReplicaInDatacenter != 4 {
		t.Fatal("info.MaxReplicaInDatacenter != 4")
	}
	if info.MaxReplicaInRack != 3 {
		t.Fatal("info.MaxReplicaInDatacenter != 3")
	}
	if info.ReplicaUsingSSD != 3 {
		t.Fatal("info.ReplicaUsingSSD != 3")
	}
}

func TestMain(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	stores := []Store{
		{1, StoreLabelInit("dc1", "rack1", "d1r1h1", ValueHDD)}, //
		{2, StoreLabelInit("dc2", "rack1", "d2r1h1", ValueHDD)},
		{7, StoreLabelInit("dc2", "rack1", "d2r1h1", ValueHDD)},
		{3, StoreLabelInit("dc3", "rack1", "d3r1h1", ValueHDD)}, //
		{4, StoreLabelInit("dc3", "rack1", "d3r1h1", ValueSSD)},
		{5, StoreLabelInit("dc3", "rack1", "d3r1h2", ValueSSD)},
		{6, StoreLabelInit("dc3", "rack2", "d3r2h3", ValueSSD)}, //
	}
	strategies := []Strategy{
		{3, 0, 0, 1, 0},
		{3, 3, 1, 0, 0},
		{3, 2, 2, 1, 0},
		{5, 2, 3, 0, 0},
		{3, 0, 0, 0, 3},
		{3, 0, 0, 0, 1},
	}

	type Plan struct {
		Idx    int
		Region []int
	}

	if false {
		plans := []Plan{
			{6, []int{7, 6, 1}},
			{3, []int{1, 3, 4, 5, 6}},
			{3, []int{5, 6, 1, 2, 3}},
			{2, []int{7, 3, 1}},
			{0, []int{5, 4, 3}},
			{3, []int{5, 3, 6, 1, 4}},
			{4, []int{2, 6, 5}},
		}
		plan := 0
		region := Region{plans[plan].Region}
		a1 := CalculateStoreInfo(NewRegionStore(stores, region), NilFilter)
		a2 := CalculateStoreInfo(NewRegionStore(stores, Region{[]int{5, 3, 6, 1, 7}}), NilFilter)
		fmt.Printf("%#v\n", strategies[plans[plan].Idx])
		println("------------ info ------------")
		fmt.Println(a1)
		_ = a2
		// fmt.Println(a2)
		println("----------- score ------------")
		// fmt.Println(a1.CalculateScore(&strategies[3]))
		// fmt.Println(a2.CalculateScore(&strategies[3]))
		testStrategy(t, stores, strategies[plans[plan].Idx], region)
		return
	}

	now := time.Now()
	const RUNTIME = 1000
	for i := 0; i < RUNTIME; i++ {
		for _, s := range strategies {
			replicas := rand.Perm(len(stores))[:s.MaxReplicas]
			for idx := range replicas {
				replicas[idx] = stores[replicas[idx]].ID
			}
			// replicas = []int{5, 3, 6, 1, 4}[:s.MaxReplicas]
			testStrategy(t, stores, s, Region{replicas})
		}
	}
	println((time.Now().Sub(now) / RUNTIME / time.Duration(len(strategies))).String())
}

func testStrategy(t *testing.T, stores []Store, s Strategy, region Region) {
	newRegion := Check(stores, region, s)
	key := []interface{}{
		"Strategy:", s, "region:", newRegion,
	}
	rs := NewRegionStore(stores, newRegion)
	info := CalculateStoreInfo(rs, NilFilter)
	if s.MaxDatacenterCnt > 0 {
		if info.DatacenterCnt > s.MaxDatacenterCnt {
			t.Fatal(key, "datacenter not match", info.DatacenterCnt, s.MaxDatacenterCnt)
		}
	}
	if s.MaxReplicaInDatacenter > 0 {
		if info.MaxReplicaInDatacenter > s.MaxReplicaInDatacenter {
			t.Fatal(key, "MaxReplicaInDatacenter not match",
				info.MaxReplicaInDatacenter, s.MaxReplicaInDatacenter)
		}
	}
	if s.MaxReplicaInRack > 0 {
		if info.MaxReplicaInRack > s.MaxReplicaInRack {
			t.Fatal(key, "MaxReplicaInRack not match",
				info.MaxReplicaInRack, s.MaxReplicaInRack)
		}
	}
	if s.MinReplicaRequireSSD > 0 {
		if info.ReplicaUsingSSD < s.MinReplicaRequireSSD {
			t.Fatal(key, "MinReplicaRequireSSD not match",
				info.ReplicaUsingSSD, s.MinReplicaRequireSSD)
		}
	}
}

func TestCalculateSwapCost(t *testing.T) {
	labels := []string{"k1", "k2"}
	{
		cost := calculateSwapCost(labels,
			&Store{ID: 1, Labels: map[string]string{"k1": "v1"}},
			&Store{ID: 2, Labels: map[string]string{"k1": "v2"}})
		if cost != 3 {
			t.Fatal("cost != 3")
		}
	}
	{
		cost := calculateSwapCost(labels,
			&Store{ID: 1, Labels: map[string]string{"k1": "v1", "k2": "v2"}},
			&Store{ID: 2, Labels: map[string]string{"k1": "v1", "k2": "v3"}})
		if cost != 2 {
			t.Fatal("cost != 2")
		}
	}
	{
		cost := calculateSwapCost(labels,
			&Store{ID: 1, Labels: map[string]string{"k1": "v1", "k2": "v2"}},
			&Store{ID: 2, Labels: map[string]string{"k1": "v1", "k2": "v2"}})
		if cost != 1 {
			t.Fatal("cost != 1")
		}
	}
	{
		cost := calculateSwapCost(labels,
			&Store{ID: 1, Labels: map[string]string{"k1": "v1", "k2": "v2"}},
			&Store{ID: 1, Labels: map[string]string{"k1": "v1", "k2": "v2"}})
		if cost != 0 {
			t.Fatal("cost != 1")
		}
	}
}
