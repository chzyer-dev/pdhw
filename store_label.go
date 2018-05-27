package pdhw

const (
	LabelDatacenter  = "dc"
	LabelRack        = "rack"
	LabelHost        = "host"
	LabelCompute     = "compute"
	LabelMemory      = "memory"
	LabelStorage     = "storage"
	LabelStorageType = "storageType"
)

const (
	ValueHDD = "hdd"
	ValueSSD = "ssd"
)

func StoreLabelInit(dc string, rack string, host string, storageType string) map[string]string {
	return map[string]string{
		LabelDatacenter:  dc,
		LabelRack:        rack,
		LabelHost:        host,
		LabelStorageType: storageType,
	}
}
