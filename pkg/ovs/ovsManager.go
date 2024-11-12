package ovs

import (
	"Netlink/api"
	"fmt"
	"github.com/digitalocean/go-openvswitch/ovs"
	"github.com/vishvananda/netlink"
	"os/exec"
	"strconv"
	"strings"
)

type OvsManager struct {
	oClinet *ovs.Client
	bridge  string
}

func NewOvsManager() *OvsManager {
	c := ovs.New()
	return &OvsManager{c, "netlink-br0"}
}

func (om *OvsManager) CreateBridge() error {
	if err := om.oClinet.VSwitch.AddBridge(om.bridge); err != nil {
		return err
	}
	return nil
}

func (om *OvsManager) DeleteBridge() error {
	if err := om.oClinet.VSwitch.DeleteBridge(om.bridge); err != nil {
		return err
	}
	return nil
}

// AddVeth adds the host side of the veth pair to the OVS bridge
func (om *OvsManager) AddVeth(vethHost string) error {
	// Ensure the veth exists
	link, err := netlink.LinkByName(vethHost)
	if err != nil {
		return fmt.Errorf("failed to find veth interface: %v", err)
	}

	// Set up the veth interface if it's not already up
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring up veth interface: %v", err)
	}

	// Add veth interface to the OVS bridge
	if err := om.oClinet.VSwitch.AddPort(om.bridge, vethHost); err != nil {
		return fmt.Errorf("failed to add veth to OVS bridge: %v", err)
	}

	return nil
}

func (om *OvsManager) AddFlowsByLink(link *api.Link) error {

	srcPort, err := GetPortId(om.bridge, link.SrcNode+"-ovs")
	if err != nil {
		return err
	}
	dstPort, err := GetPortId(om.bridge, link.DstNode+"-ovs")
	if err != nil {
		return err
	}
	if err = om.oClinet.OpenFlow.AddFlow(om.bridge, &ovs.Flow{
		Matches: []ovs.Match{
			ovs.InPortMatch(srcPort),
		},
		Actions: []ovs.Action{
			ovs.Output(dstPort),
		},
	}); err != nil {
		return err
	}
	if err = om.ApplyLinkProperty(link); err != nil {
		println("Error applying link property")
		return err
	}
	return nil
}

func (om *OvsManager) ApplyLinkProperty(link *api.Link) error {
	//if link.Properties.Rate > 0 {
	//	if err := om.SetBandwidthLimit(link.SrcNode+"-ovs", link.Properties.Rate); err != nil {
	//		return err
	//	}
	//}
	//if link.Properties.Latency > 0 {
	//	if err := om.SetLatency(link.SrcNode+"-ovs", link.Properties.Latency); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func GetPortId(bridge, port string) (int, error) {
	cmd := exec.Command("ovs-vsctl", "get", "Interface", port, "ofport")
	output, err := cmd.Output()
	if err != nil {
		return -1, fmt.Errorf("failed to get port %s id on OVS bridge %s: %v", port, bridge, err)
	}
	resultStr := strings.TrimSpace(string(output))
	resultInt, err := strconv.Atoi(resultStr)
	if err != nil {
		return -1, fmt.Errorf("error converting port %s id %s to int: %v", port, resultStr, err)
	}
	return resultInt, nil
}
