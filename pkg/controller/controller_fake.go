package controller

import (
	"consul-sync/pkg/calico"
	"consul-sync/pkg/consul"
	"consul-sync/pkg/log"
	capi "github.com/hashicorp/consul/api"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"sync"
	//"time"
)

var (
	logger = log.LoggerFunction()
)

func init() {
	log.SeValues("FakeController", "", "")
}

type SyncControllerFake struct {
	calicoC         calico.Controller
	consulC         consul.Controller
	logger          *logrus.Entry
	consulChange    map[string]int
	serviceMap      map[string][]*capi.CatalogService
	consulCalicoMap map[string]string
	mapAction       map[string]map[string]time.Time
	mutexCache      *sync.RWMutex
}

func NewFakeController(consulCalicoConf map[string]string) *SyncController {

	var consulChange = make(map[string]int)
	var serviceMap = make(map[string][]*capi.CatalogService)
	var consulCalicoMap = make(map[string]string)
	var mapAction = make(map[string]map[string]time.Time)
	mutexCache := &sync.RWMutex{}

	//Create mapping for consul service and calico globalNetworkSet
	for k, v := range consulCalicoConf {
		consulCalicoMap[k] = v
		mapAction[k] = make(map[string]time.Time)
	}

	//Create calico fake controller
	calicoC := calico.ControllerFake{
		GlobalNetworkSets: map[string]*v3.GlobalNetworkSet{},
	}

	//Create consul fake controller
	consulC := consul.ControllerFake{}

	return &SyncController{
		calicoC:         &calicoC,
		consulC:         &consulC,
		logger:          log.LoggerFunction(),
		consulChange:    consulChange,
		serviceMap:      serviceMap,
		consulCalicoMap: consulCalicoMap,
		mapAction:       mapAction,
		mutexCache:      mutexCache,
	}
}

func InitGlobalNetworkSet(c *SyncController, calicoFakeData map[string][]string) {

	var globalNetworkSets = make(map[string]*v3.GlobalNetworkSet)
	for gns := range calicoFakeData {

		//Add list of Ips to calico GlobalNetworkSet
		var globaNetworkSetSpec = v3.GlobalNetworkSetSpec{Nets: calicoFakeData[gns]}
		var globalNetworkSet = &v3.GlobalNetworkSet{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       globaNetworkSetSpec,
		}
		globalNetworkSets[gns] = globalNetworkSet
	}

	//Create calico fake controller
	var (
		calicoC = calico.ControllerFake{
			GlobalNetworkSets: globalNetworkSets,
		}
	)

	c.calicoC = &calicoC

}

func InitConsulServices(c *SyncController, consulFakeData map[string][]string) {

	for k, v := range consulFakeData {
		var catalog []*capi.CatalogService
		for _, ip := range v {
			var node capi.CatalogService
			node.Address = ip
			catalog = append(catalog, &node)
		}
		c.serviceMap[k] = catalog
	}

}
