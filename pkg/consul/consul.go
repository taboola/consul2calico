package consul

import (
	"consul-sync/pkg/log"
	"consul-sync/pkg/utils"
	"context"
	"github.com/cenkalti/backoff"
	capi "github.com/hashicorp/consul/api"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	logger 					 = log.LoggerFunction
	consulAddr  	   	     = "http://"+os.Getenv("CONSUL_ADDR")+":8500"
	consulDc   	   		     = os.Getenv("CONSUL_DC")
	consulToken 	   	     = os.Getenv("CONSUL_TOKEN")
	consulTcpTimeout,_       = time.ParseDuration(os.Getenv("CONSUL_TCP_TIMEOUT"))
	consulSyncWaitTime,_     = time.ParseDuration(os.Getenv("CONSUL_SYNC_WAIT_TIME"))
)

func init() {
	log.SeValues("consul-controller", "", "")
}

type Controller interface {
	Watch(ctx context.Context,mutexCache *sync.RWMutex,consulChange map[string]int,
		consulCalico map[string]string, serviceMap map[string][]*capi.CatalogService,stop chan os.Signal)
}

type ControllerImpl struct {
	consulClient *capi.Client
}

func New(c *capi.Client) *ControllerImpl {
	return &ControllerImpl{
		c,
	}
}


// GetConsulClient will return consul client .
func GetConsulClient() (*capi.Client,error){
	//Variables that define behavior of consul client
	defConfig := capi.DefaultConfig()
	defConfig.Transport.DialContext = (&net.Dialer{
		Timeout:   consulTcpTimeout ,
		KeepAlive: 30 * time.Second,
		DualStack: false,
	}).DialContext

	apiConf := capi.Config{
		Address:    consulAddr,
		Scheme:     "",
		Datacenter: consulDc,
		Transport:  defConfig.Transport,
		HttpClient: nil,
		HttpAuth:   nil,
		WaitTime:   0,
		Token:      consulToken,
		TokenFile:  "",
		TLSConfig:  defConfig.TLSConfig,
	}

	var consulClient *capi.Client
	logger().Infof("Starting initialize consul client")
	err := backoff.Retry(func() error{
		// Get a new client
		consulClientNew, err := capi.NewClient(&apiConf)
		consulClient = consulClientNew
		if err != nil{
			logger().Errorf("Failed to initialize consul client ERROR : %v",err)
			return err
		}
		return nil
	}, utils.GetBackOff())

	//client.
	if err != nil {
		logger().Errorf("Failed to initialize consul client 3 times ERROR : %v",err)
		return nil,err
	}else {
		logger().Info("Success to initialize consul client")
	}
	return consulClient,nil
}

// syncConsulService function will watch consul state on a given service
// Upon every change to the consul service this function will trigger a calico sync
// and will update the given GlobalNetworkSet
func watchConsulService(ctx context.Context,mutexCache *sync.RWMutex,consulChange map[string]int, consulSvc string, consulClient *capi.Client,
	serviceMap map[string][]*capi.CatalogService, stop chan os.Signal) {

	var ServiceCatalog []*capi.CatalogService
	var opts *capi.QueryOptions
	var meta *capi.QueryMeta
	var err error
	opts = &capi.QueryOptions{
		AllowStale: true,
		WaitIndex:  1,
		WaitTime:   consulSyncWaitTime,
	}


	for {
		logger().Infof("Starting sync-consul for service : %v", consulSvc)
		select {
		case <-stop:
			logger().Infof("Channel closed - Stopped sync-consul for service : %v", consulSvc)
			return
		default:
			errAfterRetry := backoff.Retry(func() error{
				// This Consul API is a blocking operation until service changes or timeout
				// https://www.consul.io/api/features/blocking
				logger().Debugf("Running Blocking query for service %v",consulSvc)
				ServiceCatalog, meta, err = consulClient.Catalog().Service(consulSvc,
					"", opts.WithContext(ctx))
				if err != nil{
					logger().Errorf("Failed sync-consul for service :  %v Time : %v  ERROR : %v",
						consulSvc,time.Now(),err)
					return err
				}
				return nil
			}, utils.GetBackOff())


			if errAfterRetry != nil {
				logger().Errorf("Failed sync-consul for service :  %v For few times . ERROR :%v",
					consulSvc,errAfterRetry)
				signal.Notify(stop, syscall.SIGTERM)
				break
			}
			//Update local serviceMap with the latest consul server list
			mutexCache.Lock()
			serviceMap[consulSvc]   = ServiceCatalog
			consulChange[consulSvc] = 1
			mutexCache.Unlock()

			opts.WaitIndex = meta.LastIndex

			//Upon every change from consul - trigger calico sync
			logger().Infof("Success - Synced accured !..... for service :  %v",consulSvc)

		}
	}

}


// Watch  Gets a list of consul services and runs a watch for each consul service.
func (c *ControllerImpl) Watch(ctx context.Context,mutexCache *sync.RWMutex,consulChange map[string]int,
	consulCalico map[string]string, serviceMap map[string][]*capi.CatalogService, stop chan os.Signal) {
	for consulService := range consulCalico {
		go watchConsulService(ctx,mutexCache,consulChange,consulService, c.consulClient, serviceMap,stop)
	}
}
