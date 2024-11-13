package link

import (
	"Netlink/api"
	"Netlink/pkg/ovs"
)

type LinkManager struct {
	om *ovs.OvsManager
}

func NewLinkManager(o *ovs.OvsManager) *LinkManager {
	return &LinkManager{
		om: o,
	}
}

func (lm *LinkManager) ApplyLink(link *api.Link, src api.Node, dst api.Node) error {
	return lm.om.AddFlowsByLink(link, src, dst)
}

// ApplyLinkProperties : Apply link properties only for unidirectional link
// directional link should be handled by the caller
func (lm *LinkManager) ApplyLinkProperties(link *api.Link, ingress *api.Node, dst api.Node) error {
	link.Properties.DstIP = dst.Interface.Ipv4
	// Check if the rule is new
	if _, existed := ingress.Rules[link.DstNode]; existed {
		return nil
	} else {
		// CreateHtbClass will modify ingress.Rules
		return lm.CreateHtbClass(link, ingress)
	}
}
