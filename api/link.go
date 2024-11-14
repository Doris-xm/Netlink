package api

type Link struct {
	Uid            int32
	SrcNode        string         `yaml:"srcNode"` // SrcNodeName
	DstNode        string         `yaml:"dstNode"` // DstNodeName
	Properties     LinkProperties `yaml:"properties"`
	UniDirectional bool           `yaml:"uniDirectional" default:"false"`

	SrcIntf NodeInterface
	DstIntf NodeInterface

	IsPhysicalVirtual bool
}

type LinkProperties struct {
	Latency       uint32  `yaml:"latency"` // in ms
	Loss          float32 `yaml:"loss"`    // in percentage
	Rate          uint64  `yaml:"rate"`    // in mbps
	HTBClassid    uint32  // netlink.Makehandle(1, 1)
	DstIP         string  // for filtering (192.168.1.1)
	NetemHandleId string
}
