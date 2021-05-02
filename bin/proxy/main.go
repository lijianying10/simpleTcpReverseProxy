package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/lijianying10/simpleTcpReverseProxy/pkg/service"
)

var serviceListenAddr *string = flag.String("l", "127.0.0.1:8080", "service listen addr")
var configFilePath *string = flag.String("c", "/etc/simpleTcpReverseProxy.json", "config file path")

func main() {
	fmt.Println("server starting....")
	flag.Parse()
	fmt.Println("Listen on :", *serviceListenAddr)
	rt := service.NewRuntime(*serviceListenAddr, *configFilePath)
	daemon.SdNotify(false, daemon.SdNotifyReady)
	go keepAlive()
	rt.Run()
}

func keepAlive() {
	interval, err := daemon.SdWatchdogEnabled(false)
	if err != nil || interval == 0 {
		return
	}
	for {
		daemon.SdNotify(false, daemon.SdNotifyWatchdog)
		time.Sleep(interval / 3)
	}
}
