package main

import (
	"flag"

	"github.com/lijianying10/simpleTcpReverseProxy/pkg/service"
)

var serviceListenAddr *string = flag.String("l", "127.0.0.1:8080", "service listen addr")

//var configFilePath *string = flag.String("c", "/etc/simpleTcpReverseProxy.json", "config file path")
var configFilePath *string = flag.String("c", "/tmp/simpleTcpReverseProxy.json", "config file path")

func main() {
	flag.Parse()
	rt := service.NewRuntime(*serviceListenAddr, *configFilePath)
	rt.Run()
}
