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
	Latency       uint32 // in ms
	Loss          float32
	Rate          uint64 // in mbps
	HTBClassid    uint32 // netlink.Makehandle(1, 1)
	DstIP         string // for filtering (192.168.1.1)
	NetemHandleId string
}
