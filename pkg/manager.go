package pkg

import (
	"Netlink/api"
	"Netlink/pkg/link"
	"Netlink/pkg/node"
	"Netlink/pkg/ovs"
	"context"
)

type Manager struct {
	Nodes map[string]api.Node
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

	m.Nodes[n.Name] = n
	err := m.cm.AddNode(m.ctx, &n)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) AddLink(l api.Link) error {
	// check src name and dst name
	l.SrcIntf = m.Nodes[l.SrcNode].Interface
	l.DstIntf = m.Nodes[l.DstNode].Interface

	print(l.SrcIntf.NodeName, l.DstIntf.NodeName)

	m.Links = append(m.Links, l)
	err := m.lm.ApplyLink(&l)
	if err != nil {
		return err
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
