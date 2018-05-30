package pdhw

import "testing"

func TestStoreStrategyInfo(t *testing.T) {
	info := &StoreStrategyInfo{
		DatacenterCnt: 1,
	}
	info.String()
}
