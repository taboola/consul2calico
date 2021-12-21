package main

import (
	"context"
	"github.com/taboola/consul2calico/pkg/calico"
	"github.com/taboola/consul2calico/pkg/consul"
	"github.com/taboola/consul2calico/pkg/controller"
	"github.com/taboola/consul2calico/pkg/log"
	"github.com/taboola/consul2calico/pkg/setting"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	logger             = log.LoggerFunction()
	servicesConfigFile = os.Getenv("SVC_CONFIG_FILE")
)

func init() {
	log.SeValues("all", "", "")
}

func main() {

	// Create application context withCancel
	ctx, cancel := context.WithCancel(context.Background())

	stop := make(chan os.Signal, 1)                                    // Create channel to receive OS signals
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the signals channel to receive SIGTERM
	mutexCache := &sync.RWMutex{}

	setting.UpdateGlobalParams(servicesConfigFile)
	var consulChange = make(map[string]int)
	var serviceMap = setting.ServiceMap
	var consulCalicoMap = setting.ConsulCalicoMap
	var mapAction = setting.MapAction

	//Init Consul client
	consulClient, err := consul.GetConsulClient()
	if err != nil {
		logger.Errorf("Failed to initialize consul client ERROR: %v", err)
		os.Exit(2)
	}

	//Init Calico client
	calicoClient, err := calico.GetCalicoClient()
	if err != nil {
		logger.Errorf("Failed to initialize calico client ERROR: %v", err)
		os.Exit(2)
	}

	SyncController := controller.New(calico.New(calicoClient), consul.New(consulClient), logger, consulChange, serviceMap,
		consulCalicoMap, mapAction, mutexCache)
	SyncController.Run(ctx, stop)

	<-stop
	// Shutdown. Cancel application context will kill all attached tasks.
	cancel()
	//Wait for goroutines to die
	time.Sleep(3 * time.Second)
	os.Exit(2)
}
