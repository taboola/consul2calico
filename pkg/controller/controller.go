package controller

import (
	"consul-sync/pkg/calico"
	"consul-sync/pkg/consul"
	"consul-sync/pkg/utils"
	"context"
	capi "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	calicoSyncInterval,_       = time.ParseDuration(utils.GetEnv("CALICO_SYNC_INTERVAL","10s"))
	calicoRemoveGraceTime,_    = time.ParseDuration(utils.GetEnv("CALICO_REMOVE_GRACE_TIME","30m"))
	controllerCheckCacheTime,_ = time.ParseDuration(utils.GetEnv("CONTROLLER_CHECK_CACHE_TIME","1s"))
)

type SyncController struct {
	calicoC calico.Controller
	consulC consul.Controller
	logger *logrus.Entry
	consulChange 	map[string]int
	serviceMap 		map[string][]*capi.CatalogService
	consulCalicoMap map[string]string
	mapAction 		map[string]map[string]time.Time
	mutexCache *sync.RWMutex
}

func New(calicoC calico.Controller, consulC consul.Controller,logger *logrus.Entry,consulChange map[string]int,
	serviceMap map[string][]*capi.CatalogService, consulCalicoMap map[string]string,
	mapAction map[string]map[string]time.Time, mutexCache *sync.RWMutex) *SyncController {
	return &SyncController{
		calicoC,
		consulC,
		logger,
		consulChange,
		serviceMap,
		consulCalicoMap,
		mapAction,
		mutexCache,
	}
}


func (c *SyncController) Run(ctx context.Context,stop chan os.Signal){

	//start sync consul
	go c.consulC.Watch(ctx,c.mutexCache,c.consulChange,c.consulCalicoMap,c.serviceMap,stop)

	//start RefreshCacheCalico
	go c.RefreshCacheCalico(stop)

	//start SyncOnTrigger
	go c.SyncOnTrigger(ctx,stop)

}

func (c *SyncController) RefreshCacheCalico(stop chan os.Signal){
	for {
		c.logger.Infof("Starting RefreshCacheCalico")
		select {
		case <-stop:
			c.logger.Infof("Channel closed - Stopped RefreshCacheCalico")
			return
		default:
			time.Sleep(calicoSyncInterval)
			// For each consulSVC given in configmap
			for consulService := range c.consulCalicoMap {
				c.AddCache(consulService)
			}
			c.logger.Infof("Success - Completed RefreshCacheCalico")
		}
	}
}

func (c *SyncController) SyncOnTrigger(ctx context.Context, stop chan os.Signal){
	for  {
		time.Sleep(controllerCheckCacheTime)
		for consulService := range c.GetServicesChanged(){
			c.logger.Debugf("Running triggered Sync for Service %v",consulService)
			c.SyncGlobalNetworkSet(ctx,consulService,c.consulCalicoMap[consulService],stop)
			c.DeleteCache(consulService)
		}
	}
}


func (c *SyncController) SyncGlobalNetworkSet(ctx context.Context,consulService string,
	calicoGNetworkSet string, stop chan os.Signal)  {

	c.mutexCache.Lock()

	c.logger.Infof("Starting SyncGlobalNetworkSet for : %v",calicoGNetworkSet)
	ConsulServiceCatalog := c.GetCatalogService(consulService)
	//Trigger PD/LOG if consul service is Empty :
	if len(ConsulServiceCatalog) == 0 {
		c.logger.Errorf("Consul2Calico Error - consul service empty : %v",consulService)
	}

	//Get consul servers list
	var ipsConsul []string
	for _, node := range ConsulServiceCatalog {
		ipsConsul = append(ipsConsul, node.Address+"/32")
	}

	c.logger.Debugf("Num of Servers in consul For service %v: %v",
		consulService, len(ipsConsul))

	//Check if server is in consul and Remove from mapAction - For example rebuild/reboot
	for _, value := range ipsConsul {
		delete(c.mapAction[consulService], value)
	}

	//Get calico server list
	net, Err := c.calicoC.GetNetworkSet(ctx, calicoGNetworkSet)
	if Err != nil {
		c.logger.Errorf("Failed to getNetworkSet(%v) Error: %v",calicoGNetworkSet,Err)
		c.mutexCache.Unlock()
		signal.Notify(stop, syscall.SIGTERM)
		//break
	}

	ipsCalico := net.Spec.Nets

	//Get difference between state in consul and state in calico
	a , b := utils.CompareSlice(ipsConsul,ipsCalico)

	//Compare Consul - Calico
	c.logger.Debugf("Num of Servers in Consul but not in calico For service %v: %v",
		consulService, len(a))
	//Add servers (In consul , Not In calico) to mapAction
	for _, value := range a {
		c.mapAction[consulService][value] = time.Time{}.UTC()
	}

	//Compare Calico - Consul
	c.logger.Debugf("Num of Servers in Calico but not in consul For service %v: %v",
		consulService, len(b))
	//Add servers (In calico , Not In consul) to mapAction
	for _, value := range b {
		_, exists := c.mapAction[consulService][value]
		if !exists {
			c.mapAction[consulService][value] = time.Now().UTC()
		}
	}

	ipsAddCalico,ipsDeleteCalico := c.mapLoopAction(consulService)

	//Run update only if needed
	if len(ipsAddCalico) + len(ipsDeleteCalico) > 0{
		err := c.calicoC.UpdateNetworkSet(ctx,calicoGNetworkSet,ipsAddCalico,ipsDeleteCalico,stop)
		if err != nil {
			c.logger.Errorf("Failed to UpdateNetworkSet(%v) Error: %v",calicoGNetworkSet,err)
			signal.Notify(stop, syscall.SIGTERM)
		}
	}
	c.mutexCache.Unlock()
	c.logger.Infof("Success - Completed SyncGlobalNetworkSet")
}


func (c *SyncController) mapLoopAction(consulService string)([]string,[]string){
	var ipsAddCalico []string
	var ipsDeleteCalico []string

	//Run on servers in mapAction ,modify calico GlobalNetworkSet based on timestamps
	for ipAddr, timestamp := range c.mapAction[consulService] {
		if timestamp.IsZero() {
			c.logger.Infof("Add to bulk API call - Add server(%v) to calico For service %v",
				ipAddr, consulService)
			//Add server to Add list
			ipsAddCalico = append(ipsAddCalico,ipAddr)
			delete(c.mapAction[consulService], ipAddr)
		} else if timestamp.Add(calicoRemoveGraceTime).Before(time.Now().UTC()) {
			c.logger.Infof("Add to bulk API call - Remove server(%v) from calico For service %v",
				ipAddr, consulService)
			//Add server to Remove list
			ipsDeleteCalico = append(ipsDeleteCalico,ipAddr)
			delete(c.mapAction[consulService], ipAddr)
		} else {
			//Wait for server timestamp to reach
			c.logger.Infof("Waiting to Remove server %v from GlobalNetworkSet %v timestamp: %v",
				ipAddr,c.consulCalicoMap[consulService], timestamp)
		}
	}

	return ipsAddCalico,ipsDeleteCalico

}

func (c *SyncController) GetServicesChanged()map[string]int{
	return c.consulChange
}

func (c *SyncController) DeleteCache(consulService string){
	c.mutexCache.Lock()
	delete(c.consulChange, consulService)
	c.mutexCache.Unlock()
}

func (c *SyncController) AddCache(consulService string){
	c.mutexCache.Lock()
	c.consulChange[consulService] = 1
	c.mutexCache.Unlock()
}

func (c *SyncController) GetCatalogService(consulService string)[]*capi.CatalogService{
	return c.serviceMap[consulService]
}