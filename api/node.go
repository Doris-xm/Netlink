package api

type Node struct {
	Uid       int32
	Name      string
	Interface NodeInterface
	NetNs     string
	IsNormal  bool
	Image     string

	Rules map[string]LinkProperties // len(Rules) will never decrease, used for classid
}

type NodeInterface struct {
	Uid      int32
	Name     string
	Mac      string
	Ipv4     string
	Ipv6     string
	NetNs    string
	Class    string
	NodeName string
	BrName   string
}
