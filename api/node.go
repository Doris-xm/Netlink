package api

type Node struct {
	Uid       int32
	Name      string
	Interface NodeInterface
	NetNs     string
	IsNormal  bool
	Image     string
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
