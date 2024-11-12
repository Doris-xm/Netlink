package api

type Link struct {
	Uid            int32
	SrcNode        string
	DstNode        string
	Properties     LinkProperties
	UniDirectional bool

	SrcIntf NodeInterface
	DstIntf NodeInterface

	IsPhysicalVirtual bool
}

type LinkProperties struct {
	Latency       string
	LatencyCorr   string
	Jitter        string
	Loss          string
	LossCorr      string
	Rate          string
	Gap           uint32
	Duplicate     string
	DuplicateCorr string
	ReorderProb   string
	ReorderCorr   string
	CorruptProb   string
	CorruptCorr   string
}
