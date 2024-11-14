package main

import (
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
		// 循环接收并执行命令
		for {
			var input string
			fmt.Print("Enter command: ")
			_, err := fmt.Scanln(&input)
			if err != nil {
				fmt.Println("Error reading input:", err)
				continue
			}

			// 根据输入的命令执行相应的操作
			switch input {
			case "apply":
				_, err := fmt.Scanln(&input)
				if err != nil {
					fmt.Println("Error reading input:", err)
					return
				}
				err = c.ApplyTopoConfig(input)
				if err != nil {
					stop <- syscall.SIGTERM
					fmt.Println("Error applying configuration:", err)
				} else {
					fmt.Println("Configuration applied successfully.")
				}
			case "exit":
				fmt.Println("Exiting...")
				stop <- syscall.SIGTERM
				return
			case "show":
				_, err := fmt.Scanln(&input)
				if err != nil {
					fmt.Println("Error reading input:", err)
					return
				}
				switch input {
				case "nodes":
					c.ShowNodes()
				case "links":
					c.ShowLinks()
				}
			default:
				continue
			}
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
	//c.Destroy()

}
