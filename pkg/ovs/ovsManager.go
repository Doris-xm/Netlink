package ovs

import (
	"Netlink/api"
	"fmt"
	"github.com/digitalocean/go-openvswitch/ovs"
	"github.com/sirupsen/logrus"
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
	// 删除默认的 NORMAL 规则
	if err := om.oClinet.OpenFlow.DelFlows(om.bridge, &ovs.MatchFlow{}); err != nil {
		return err
	}

	// sudo ovs-vsctl set bridge ovs-br-host datapath_type=system
	cmd := exec.Command("ovs-vsctl", "set", "bridge", "netlink-br0", "datapath_type=system")
	err := cmd.Run()
	if err != nil {
		logrus.Fatalf("error setting OVS bridge %s datapath type: %v", "netlink-br0", err)
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

func (om *OvsManager) AddFlowsByLink(link *api.Link, src api.Node, dst api.Node) error {
	// Add flow to ovs group table
	//  ovs-ofctl mod-group netlink-br0 group_id=2,type=all,bucket=output:"node1-ovs",bucket=output:"node3-ovs"
	cmd := exec.Command("ovs-ofctl", "mod-group", om.bridge, "group_id="+strconv.Itoa(src.Uid)+",type=all,bucket=output:\""+dst.Name+"-ovs\"")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add group table: %v", err)
	}

	cmd = exec.Command("ovs-ofctl", "mod-group", om.bridge, "group_id="+strconv.Itoa(dst.Uid)+",type=all,bucket=output:\""+src.Name+"-ovs\"")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add group table: %v", err)
	}

	return err
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

func (om *OvsManager) AddGroupTable(intf string, groupId int) error {
	// ovs-ofctl add-group netlink-br0 group_id=2,type=all
	cmd := exec.Command("ovs-ofctl", "add-group", om.bridge, "group_id="+strconv.Itoa(groupId)+",type=all")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add group table: %v", err)
	}

	in_port, _ := GetPortId(om.bridge, intf)
	//ovs-ofctl add-flow netlink-br0 in_port=7,actions=group:2
	cmd = exec.Command("ovs-ofctl", "add-flow", om.bridge, "in_port="+strconv.Itoa(in_port)+",actions=group:"+strconv.Itoa(groupId))
	err = cmd.Run()

	return err
}
