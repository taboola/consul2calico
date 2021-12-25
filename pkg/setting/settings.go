package setting

import (
	capi "github.com/hashicorp/consul/api"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

// ServiceMap holds the full Catalog of the service in consul
var ServiceMap = make(map[string][]*capi.CatalogService)

// ConsulCalicoMap mapping of consul service and a matched GlobalNetworkSet
var ConsulCalicoMap = make(map[string]string)

// MapAction is a map of Ip addresses that needs to be deleted or added to Calico
var MapAction = make(map[string]map[string]time.Time)

// UpdateGlobalParams updates parameters based on given yaml file
func UpdateGlobalParams(path string) {

	yfile, err := ioutil.ReadFile(path)

	if err != nil {

		log.Fatal(err)
	}

	data := make(map[string]string)

	err2 := yaml.Unmarshal(yfile, &data)

	if err2 != nil {

		log.Fatal(err2)
	}

	for k, v := range data {
		ConsulCalicoMap[k] = v
		MapAction[k] = make(map[string]time.Time)
	}

}
