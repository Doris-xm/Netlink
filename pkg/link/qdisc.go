package link

import (
	"Netlink/api"
	"fmt"
	ns "github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
)

// 1. Only finish htb qdisc
// 2. Use 1:0 as all parent handle
// 3. Filter by destination IP

// CreateRootQdisc : tc qdisc add dev eth0 root handle 1: htb default 30
func (lm *LinkManager) CreateRootQdisc(n api.Node) error {
	// enter container namespace
	containerNs, err := ns.GetNS(n.NetNs)
	if err != nil {
		return fmt.Errorf("failed to get namespace for container: %v", err)
	}
	defer containerNs.Close()

	err = containerNs.Do(func(_ ns.NetNS) error {
		// get link by name
		link, err := netlink.LinkByName(n.Name + "-veth0")
		if err != nil {
			return fmt.Errorf("failed to get link by name: %v", err)
		}

		// set HTB root qdisc
		qdisc := netlink.NewHtb(
			netlink.QdiscAttrs{
				LinkIndex: link.Attrs().Index,
				Handle:    netlink.MakeHandle(1, 0), // root qdisc
				Parent:    netlink.HANDLE_ROOT,      // root handle
			})
		qdisc.Defcls = 1 // Default classid 1:1

		// add HTB root qdisc
		if err := netlink.QdiscAdd(qdisc); err != nil {
			return fmt.Errorf("failed to add HTB root qdisc: %v", err)
		}
		return nil
	})

	return err
}

// CreateHtbClass :
// tc class add dev eth0 parent 1: classid 1:2 htb rate 1mbit burst 10000
// tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip dst 192.168.1.1 flowid 1:1
// tc qdisc add dev eth0 parent 1:1 handle 10: netem delay 100ms
// will modify node.Rules, record the classid
// bw control comes before loss and latency
func (lm *LinkManager) CreateHtbClass(l *api.Link, n *api.Node) error {

	if l.Properties.Latency <= 0 && l.Properties.Rate <= 0 && l.Properties.Loss <= 0 {
		return nil
	}
	l.Properties.HTBClassid = netlink.MakeHandle(1, uint16(len(n.Rules)+2)) // +2 for root and default class
	n.Rules[l.DstNode] = l.Properties

	// enter container namespace
	containerNs, err := ns.GetNS(n.NetNs)
	if err != nil {
		return fmt.Errorf("failed to get namespace for container: %v", err)
	}
	defer containerNs.Close()

	err = containerNs.Do(func(_ ns.NetNS) error {
		// get link by name
		link, err := netlink.LinkByName(n.Name + "-veth0")
		if err != nil {
			return fmt.Errorf("failed to get link by name: %v", err)
		}

		// 1. bw control
		class := netlink.NewHtbClass(
			netlink.ClassAttrs{
				LinkIndex: link.Attrs().Index,
				Handle:    l.Properties.HTBClassid,  // classid 1:2
				Parent:    netlink.MakeHandle(1, 0), // parent 1:
			},
			netlink.HtbClassAttrs{
				Rate:   l.Properties.Rate * 1024 * 1024, // rate 1mbit
				Buffer: 10000,                           // burst 10000
				Prio:   1,
			},
		)

		if err := netlink.ClassAdd(class); err != nil {
			return fmt.Errorf("failed to add HTB class: %v", err)
		}

		//// 2. filter by destination IP
		//ipInt, err := IpToInt(l.Properties.DstIP)
		//if err != nil {
		//	return err
		//}

		//filter := &netlink.U32{
		//	FilterAttrs: netlink.FilterAttrs{
		//		LinkIndex: link.Attrs().Index,
		//		Parent:    netlink.MakeHandle(1, 0),
		//		Handle:    l.Properties.HTBClassid,
		//		Priority:  1,
		//	},
		//	// Match src IP 192.168.1.1
		//	Sel: &netlink.TcU32Sel{
		//		Keys: []netlink.TcU32Key{
		//			{
		//				Mask:    0xffffffff,
		//				Val:     ipInt, // Using the converted IP integer
		//				Off:     12,    // Offset for source IP in IP header
		//				OffMask: 0,
		//			},
		//		},
		//	},
		//	ClassId: l.Properties.HTBClassid,
		//}
		//
		//if err := netlink.FilterAdd(filter); err != nil {
		//	return fmt.Errorf("failed to add u32 filter: %v", err)
		//}
		return nil
	})

	return err
}

func IpToInt(IP string) (uint32, error) {
	if strings.Contains(IP, "/") {
		IP = strings.Split(IP, "/")[0]
	}
	// Parse the source IP string to net.IP format
	ip := net.ParseIP(IP)
	if ip == nil {
		return 0, fmt.Errorf("invalid IP address: %v", IP)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, fmt.Errorf("only IPv4 addresses are supported")
	}

	// Convert the IP to integer format for the U32 filter
	ipInt := (uint32(ip4[0]) << 24) | (uint32(ip4[1]) << 16) | (uint32(ip4[2]) << 8) | uint32(ip4[3])
	return ipInt, nil

}
