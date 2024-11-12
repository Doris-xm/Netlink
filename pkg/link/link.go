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

func (lm *LinkManager) ApplyLink(link *api.Link) error {
	return lm.om.AddFlowsByLink(link)
}
