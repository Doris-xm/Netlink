package pkg

import (
	"Netlink/api"
	"Netlink/pkg/link"
	"Netlink/pkg/node"
	"Netlink/pkg/ovs"
	"context"
	"fmt"
)

// Manager handles the management of nodes, links, and network configurations
// in the system. It is responsible for adding nodes, linking nodes, applying
// link properties, and cleaning up resources when destroyed.
type Manager struct {
	Nodes map[string]api.Node // map node name to node
	om    *ovs.OvsManager
	lm    *link.LinkManager
	cm    *node.ContainerManager
	ctx   context.Context
}

// NewManager creates a new Manager instance with the default OVS manager
// and container manager.
func NewManager() *Manager {
	om := ovs.NewOvsManager()
	cm := node.NewContainerManager(om)
	lm := link.NewLinkManager(om)

	return &Manager{
		Nodes: make(map[string]api.Node),
		om:    om,
		lm:    lm,
		cm:    cm,
		ctx:   context.Background(),
	}
}

func (m *Manager) AddNode(n api.Node) error {

	// Initialize
	if n.Rules == nil {
		n.Rules = make(map[string]api.LinkProperties)
	}

	// check if existed
	if _, existed := m.Nodes[n.Name]; existed {
		if err := m.cm.DeleteNode(m.ctx, &n); err != nil {
			return err
		}
	}

	m.Nodes[n.Name] = n
	err := m.cm.AddNode(m.ctx, &n)
	if err != nil {
		return err
	}
	if err = m.lm.CreateRootQdisc(n); err != nil {
		return err
	}
	// update node
	m.Nodes[n.Name] = n

	return nil
}

func (m *Manager) AddLink(l api.Link) error {
	// check invalid link
	if _, existed := m.Nodes[l.SrcNode]; !existed {
		return fmt.Errorf("src node %s not found", l.SrcNode)
	}
	if _, existed := m.Nodes[l.DstNode]; !existed {
		return fmt.Errorf("dst node %s not found", l.DstNode)
	}

	// check src name and dst name
	src := m.Nodes[l.SrcNode]
	dst := m.Nodes[l.DstNode]
	l.SrcIntf = src.Interface
	l.DstIntf = dst.Interface

	// check if existed
	if _, existed := src.Rules[l.DstNode]; !existed {
		src.Rules[l.DstNode] = api.LinkProperties{}
		if err := m.lm.ApplyLink(src); err != nil {
			return err
		}
	}

	// directional link
	if _, existed := dst.Rules[l.SrcNode]; !existed {
		dst.Rules[l.SrcNode] = api.LinkProperties{}
		if err := m.lm.ApplyLink(dst); err != nil {
			return err
		}
	}

	// Apply Properties
	if err := m.lm.ApplyLinkProperties(&l, &src, dst); err != nil {
		return err
	}

	if !l.UniDirectional {
		if err := m.lm.ApplyLinkProperties(&l, &dst, src); err != nil {
			return err
		}

	}
	// update src & dst node
	m.Nodes[l.SrcNode] = src
	m.Nodes[l.DstNode] = dst

	return nil
}

func (m *Manager) Destroy() {
	for _, n := range m.Nodes {
		err := m.cm.DeleteNode(m.ctx, &n)
		if err != nil {
			println(err.Error())
		}
	}
	err := m.om.DeleteBridge()
	if err != nil {
		println(err.Error())
		return
	}
}
