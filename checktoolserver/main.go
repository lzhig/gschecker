package main

import (
	"fmt"
	"globaltedinc/framework/network"
	"os"
	"strconv"

	"globaltedinc/framework/glog"
)

func printHelp() {
	fmt.Println("Usage: exe addr maxusers")
	fmt.Println("params:")
	fmt.Println("addr     - TCP address")
	fmt.Println("maxusers - the number of max clients")
}

func main() {
	if len(os.Args) < 3 {
		printHelp()
		return
	}
	var q chan byte

	addr := os.Args[1]
	maxUsers, err := strconv.ParseUint(os.Args[2], 10, 32)
	if err != nil {
		printHelp()
		return
	}

	var s network.TCPServer
	s.Start(addr, uint32(maxUsers),
		func(conn *network.Connection) {
			glog.Info("client " + conn.RemoteAddr() + " connected")
		},

		func(conn *network.Connection, err error) {
			glog.Info("client " + conn.RemoteAddr() + " disconnected")
			glog.Info(err)
		},

		func(conn *network.Connection, packet *network.Packet) {
			s.SendPacket(conn, packet)
		})

	<-q
}
