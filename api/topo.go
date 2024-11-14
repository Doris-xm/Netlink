package api

type TopoConfig struct {
	Nodes []Node `yaml:"nodes"`
	Links []Link `yaml:"links"`
}
