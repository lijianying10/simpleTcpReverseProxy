package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"

	"github.com/julienschmidt/httprouter"
	"github.com/lijianying10/simpleTcpReverseProxy/pkg/config"
	"github.com/lijianying10/simpleTcpReverseProxy/pkg/proxy"
)

type ProxyRuntimeBoundle struct {
	cfg *config.Config
	rt  *proxy.Runtime
}

type Runtime struct {
	proxy                  map[string]ProxyRuntimeBoundle // name -> proxy
	listen, configFilePath string
}

func NewRuntime(listen, configFilePath string) *Runtime {
	rt := &Runtime{
		proxy:          map[string]ProxyRuntimeBoundle{},
		listen:         listen,
		configFilePath: configFilePath,
	}
	cfgList, err := config.ConfigLoader(rt.configFilePath)
	if err != nil {
		fmt.Println("[ERROR] cannot load config from file:", rt.configFilePath)
	}
	for _, cfg := range cfgList {
		if err := cfg.Valid(); err != nil {
			fmt.Println("[ERROR]: ", cfg.Name, "not valid", err.Error())
			continue
		}
		proxy := proxy.NewRuntime(cfg)
		rt.proxy[cfg.Name] = ProxyRuntimeBoundle{
			cfg: cfg,
			rt:  proxy,
		}
		go proxy.Run()
	}
	return rt
}

func (rt *Runtime) Run() {
	router := httprouter.New()
	router.GET("/ping", rt.Ping)
	router.GET("/stat", rt.Stat)
	router.GET("/cfg/get", rt.GetCurrentConfig)
	router.PUT("/cfg/set", rt.SetCurrentConfig)
	log.Fatal(http.ListenAndServe(rt.listen, router))
}

func (rt *Runtime) Stat(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "num goroutine: %d\n", runtime.NumGoroutine())
}

func (rt *Runtime) Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Pong!")
}

func (rt *Runtime) GetCurrentConfig(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var cfgs []*config.Config
	for _, rt := range rt.proxy {
		cfgs = append(cfgs, rt.cfg)
	}
	body, err := json.MarshalIndent(cfgs, " ", " ")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Write(body)
}

func (rt *Runtime) SetCurrentConfig(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("error read request body" + err.Error()))
		return
	}
	defer r.Body.Close()
	var res []*config.Config
	err = json.Unmarshal(body, &res)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("error read json body" + err.Error()))
		return
	}
	// Do some CURD work
	var report string
	for _, cfg := range res {
		if val, ok := rt.proxy[cfg.Name]; ok {
			if err := cfg.Valid(); err != nil {
				fmt.Println("[ERROR]: ", cfg.Name, "not valid", err.Error())
				continue
			}
			if !val.cfg.Same(cfg) {
				// update config
				val.rt.Stop()
				val.cfg = cfg
				val.rt = proxy.NewRuntime(cfg)
				go val.rt.Run()
				rt.proxy[cfg.Name] = val
				report += fmt.Sprintf("UPDATE %s\n", cfg.Name)
			}
		} else {
			// create config
			prt := proxy.NewRuntime(cfg)
			go prt.Run()
			rt.proxy[cfg.Name] = ProxyRuntimeBoundle{
				cfg: cfg,
				rt:  prt,
			}
			report += fmt.Sprintf("CREATE %s\n", cfg.Name)
		}
	}

	NameExistInRes := func(name string) bool {
		for _, cfg := range res {
			if cfg.Name == name {
				return true
			}
		}
		return false
	}

	for name, rtb := range rt.proxy {
		if !NameExistInRes(name) {
			// Delete runtime
			rtb.rt.Stop()
			delete(rt.proxy, name)
			report += fmt.Sprintf("DELETE %s\n", name)
		}
	}
	err = config.ConfigStore(rt.configFilePath, res)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(report + "error save config to file" + err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(report))
}
