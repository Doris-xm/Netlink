package api

type Node struct {
	Uid       int
	Name      string        `yaml:"name"`
	Interface NodeInterface `yaml:"interface"`
	NetNs     string
	IsNormal  bool
	Image     string `yaml:"image"`

	Rules map[string]LinkProperties // len(Rules) will never decrease, used for classid, map dst --> properties
}

type NodeInterface struct {
	Uid      int32
	Name     string
	Mac      string
	Ipv4     string `yaml:"ipv4"`
	Ipv6     string
	NetNs    string
	Class    string
	NodeName string
	BrName   string
}
