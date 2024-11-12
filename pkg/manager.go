package pkg

import (
	"Netlink/api"
	"Netlink/pkg/link"
	"Netlink/pkg/node"
	"Netlink/pkg/ovs"
	"context"
	"fmt"
)

type Manager struct {
	Nodes map[string]api.Node // map node name to node
	Links []api.Link
	om    *ovs.OvsManager
	lm    *link.LinkManager
	cm    *node.ContainerManager
	ctx   context.Context
}

func NewManager() Manager {
	om := ovs.NewOvsManager()
	err := om.CreateBridge()
	if err != nil {
		panic(err)
	}
	cm := node.NewContainerManager(om)
	lm := link.NewLinkManager(om)

	return Manager{
		Nodes: make(map[string]api.Node),
		Links: make([]api.Link, 0),
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

	m.Nodes[n.Name] = n
	err := m.cm.AddNode(m.ctx, &n)
	if err != nil {
		return err
	}
	if err = m.lm.CreateRootQdisc(n); err != nil {
		return err
	}
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
	l.SrcIntf = m.Nodes[l.SrcNode].Interface
	l.DstIntf = m.Nodes[l.DstNode].Interface

	m.Links = append(m.Links, l)
	err := m.lm.ApplyLink(&l)
	if err != nil {
		return err
	}

	// Apply Properties
	src := m.Nodes[l.SrcNode]
	dst := m.Nodes[l.DstNode]

	if err = m.lm.ApplyLinkProperties(&l, &src, dst); err != nil {
		return err
	}

	// update src node
	m.Nodes[l.SrcNode] = src

	if !l.UniDirectional {
		if err = m.lm.ApplyLinkProperties(&l, &dst, src); err != nil {
			return err
		}

		m.Nodes[l.DstNode] = dst
	}

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
