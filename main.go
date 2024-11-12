package main

import (
	"Netlink/api"
	"Netlink/pkg"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// 需求：模拟高延迟高带宽的链路
// 实现：通过netlink库实现，不使用现有qdisc逻辑
// 连接：起n个docker，每个docker通过veth连接到ovs交换机，ovs交换机通过veth连接到host
// 函数：SetupNode，SetupLinks，SetupOvs，ConfigLinks

func main() {
	m := pkg.NewManager()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	for i := 1; i <= 3; i++ {
		err := m.AddNode(api.Node{
			Name:  "node" + strconv.Itoa(i),
			Image: "frr:v4",
			Interface: api.NodeInterface{
				Ipv4: "192.168.1." + strconv.Itoa(i) + "/24",
			},
		})
		if err != nil {
			println(err.Error())
			return
		}
	}

	err := m.AddLink(api.Link{
		SrcNode: "node1",
		DstNode: "node2",
	})
	if err != nil {
		println(err.Error())
		return
	}
	err = m.AddLink(api.Link{
		SrcNode: "node3",
		DstNode: "node2",
	})
	if err != nil {
		println(err.Error())
		return
	}

	// wait, before shutting down , clear up the resources
	<-stop
	m.Destroy()

}
