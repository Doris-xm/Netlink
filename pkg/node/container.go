package node

import (
	"Netlink/api"
	"Netlink/pkg/ovs"
	"Netlink/pkg/util"
	"context"
	"fmt"
	ns "github.com/containernetworking/plugins/pkg/ns"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/vishvananda/netlink"
	"net"
)

const (
	NodeVethSuffix = "-veth0"
	DefaultImage   = "frr:v4"
)

// ContainerManager manages the lifecycle of containers
// seq is used to assign a unique id to each container( for ovs group id)
// seq will never decrease
type ContainerManager struct {
	dClient *client.Client
	om      *ovs.OvsManager
	seq     int
}

func NewContainerManager(o *ovs.OvsManager) *ContainerManager {
	dClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		println("Error creating docker client")
	}
	return &ContainerManager{
		dClient: dClient,
		om:      o,
		seq:     1,
	}
}

// AddNode creates a container with the given node configuration
// start the container, get NetNS
// link the container to ovs bridge
func (cm *ContainerManager) AddNode(ctx context.Context, n *api.Node) error {
	n.Uid = cm.seq
	cm.seq++
	// check illegal
	if n.Image == "" {
		n.Image = DefaultImage
	}
	if !util.CheckInvalidIpv4(n.Interface.Ipv4) {
		n.Interface.Ipv4 = "192.168.10." + fmt.Sprintf("%d", n.Uid) + "/24"

		println("node ", n.Name, " has empty or invalid or system-reserved ipv4 address,"+
			" reset to "+n.Interface.Ipv4)
	}

	// Create the container
	sysctls := make(map[string]string)
	sysctls["net.ipv4.ip_forward"] = "1"
	sysctls["net.ipv6.conf.all.forwarding"] = "1"

	_, err := cm.dClient.ContainerCreate(ctx, &container.Config{
		Image:           n.Image,
		NetworkDisabled: true,
		User:            "root",
	}, &container.HostConfig{
		Privileged: true,
		Binds:      []string{},
		Sysctls:    sysctls,
	}, nil, nil, n.Name)
	if err != nil {
		println("Error creating container")
		return err
	}

	err = cm.dClient.ContainerStart(ctx, n.Name, container.StartOptions{})
	if err != nil {
		println("Error starting container")
		return err
	}

	// Get Ns from container
	res, err := cm.dClient.ContainerInspect(ctx, n.Name)
	if err != nil {
		println("Error inspecting container")
		return err
	}
	n.NetNs = fmt.Sprintf("/proc/%d/ns/net", res.State.Pid)
	println("NetNs: ", n.NetNs)

	return cm.LinkNodeToOVS(n)

}

// LinkNodeToOVS links the container to the OVS bridge
func (cm *ContainerManager) LinkNodeToOVS(n *api.Node) error {
	err := cm.CreateVethPair(n)
	if err != nil {
		return err
	}

	// Create group table for nodex-ovs
	err = cm.om.AddGroupTable(n.Name+"-ovs", n.Uid)
	return err
}

// CreateVethPair creates a veth pair and moves one end to the container
// and adds the other end to the OVS bridge
// add ipv4 address to the container end
func (cm *ContainerManager) CreateVethPair(n *api.Node) error {

	// 1. Create Veth pair
	vethContainer := n.Name + NodeVethSuffix
	vethOvs := n.Name + ovs.VethOvsSideSuffix
	linkAttr := netlink.NewLinkAttrs()
	linkAttr.Name = vethOvs
	linkAttr.MTU = 1500
	linkAttr.Flags = net.FlagUp

	veth0 := &netlink.Veth{
		LinkAttrs: linkAttr,
		PeerName:  vethContainer,
	}

	err := netlink.LinkAdd(veth0)
	if err != nil {
		println("Error creating veth pair")
		return err
	}

	// 2. Bring up the veth pair
	containerLink, err := netlink.LinkByName(vethContainer)
	if err != nil {
		println("Error getting link by name")
		return err
	}

	if err = netlink.LinkSetUp(containerLink); err != nil {
		println("Error setting link up: ", containerLink.Attrs().Name)
		return err
	}

	hostLink, err := netlink.LinkByName(vethOvs)
	if err != nil {
		println("Error getting link by name")
		return err
	}
	if err = netlink.LinkSetUp(hostLink); err != nil {
		println("Error setting link up: ", hostLink.Attrs().Name)
		return err
	}

	// 3. Move one end to container

	// find the container network namespace
	containerNs, err := ns.GetNS(n.NetNs)
	if err != nil {
		return fmt.Errorf("failed to get namespace for container: %v", err)
	}
	defer containerNs.Close()

	if err = netlink.LinkSetNsFd(containerLink, int(containerNs.Fd())); err != nil {
		return fmt.Errorf("failed to set namespace for veth: %v", err)
	}

	// 3. Add addr ipv4
	if err = containerNs.Do(func(_ ns.NetNS) error {
		// get the link in the container namespace
		containerVeth, err := netlink.LinkByName(vethContainer)
		if err != nil {
			return fmt.Errorf("failed to get link in container namespace: %v", err)
		}

		ip, ipNet, err := net.ParseCIDR(n.Interface.Ipv4)
		if err != nil {
			return fmt.Errorf("failed to parse CIDR: %v", err)
		}
		if err = netlink.AddrAdd(containerVeth, &netlink.Addr{IPNet: &net.IPNet{IP: ip, Mask: ipNet.Mask}}); err != nil {
			return fmt.Errorf("failed to add address to link: %v", err)
		}

		// bring the link up
		if err = netlink.LinkSetUp(containerLink); err != nil {
			return fmt.Errorf("failed to set link up: %v", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to configure container namespace: %v", err)
	}

	// 4. Connect to OVS
	if err = cm.om.AddVeth(vethOvs); err != nil {
		println("Error adding veth to OVS")
		return err
	}

	// 5. record veth information
	n.Interface.Mac = veth0.Attrs().HardwareAddr.String()
	n.Interface.Name = vethContainer
	n.Interface.NodeName = n.Name
	return nil
}

func (cm *ContainerManager) DeleteNode(ctx context.Context, n *api.Node) error {

	err := cm.dClient.ContainerRemove(ctx, n.Name, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}
