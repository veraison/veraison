package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/veraison/veraison/provisioning/api"
	"github.com/veraison/veraison/provisioning/decoder"
	"github.com/veraison/veraison/provisioning/storeclient"
)

// TODO(tho) make these configurable
const (
	PluginDir  = "../plugins/bin/"
	ListenAddr = "localhost:8888"
)

func main() {
	pluginManager := NewGoPluginManager(PluginDir)
	storeClient := storeclient.NewDummy()
	apiHandler := api.NewHandler(pluginManager, storeClient)
	go apiServer(apiHandler, ListenAddr)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go terminator(sigs, done, pluginManager)
	<-done
	log.Println("bye!")
}

func terminator(sigs chan os.Signal, done chan bool, pluginManager decoder.IDecoderManager) {
	sig := <-sigs
	log.Println(sig, "received, exiting")
	if err := pluginManager.Term(); err != nil {
		log.Println("plugin manager termination failed:", err)
	}
	done <- true
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}

func NewGoPluginManager(dir string) decoder.IDecoderManager {
	mgr := &decoder.GoPluginDecoderManager{}
	err := mgr.Init(dir)
	if err != nil {
		log.Fatalf("plugin initialisation failed: %v", err)
	}

	return mgr
}
