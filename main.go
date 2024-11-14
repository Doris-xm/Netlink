package main

import (
	"Netlink/cmd"
	"Netlink/pkg"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// 需求：模拟高延迟高带宽的链路
// 实现：通过netlink库实现，不使用现有qdisc逻辑
// 连接：起n个docker，每个docker通过veth连接到ovs交换机，ovs交换机通过veth连接到host
// 函数：SetupNode，SetupLinks，SetupOvs，ConfigLinks

var c *pkg.Calculator

func main() {
	c = pkg.NewCalculator()

	defer c.Destroy()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := cmd.Execute(c); err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}
	}()

	//err := c.ApplyTopoConfig("example/topo.yaml")
	//if err != nil {
	//	c.Destroy()
	//	log.Fatal(err.Error())
	//	return
	//}
	// wait, before shutting down , clear up the resources
	<-stop

}
