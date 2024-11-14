package pkg

import (
	"Netlink/api"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Calculator struct {
	m *Manager
}

func NewCalculator() *Calculator {
	return &Calculator{
		m: NewManager(),
	}
}

func (c *Calculator) ApplyTopoConfig(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}

	// Unmarshal the YAML file into api.TopoConfig
	var topoCfg api.TopoConfig
	if err = yaml.Unmarshal(data, &topoCfg); err != nil {
		return fmt.Errorf("error unmarshaling YAML file: %v", err)
	}

	// Add nodes
	for _, node := range topoCfg.Nodes {
		if err = c.m.AddNode(node); err != nil {
			return err
		}
	}

	// Add links
	for _, link := range topoCfg.Links {
		if err = c.m.AddLink(link); err != nil {
			return err
		}
	}

	return nil
}

func (c *Calculator) Destroy() {
	c.m.Destroy()
}
