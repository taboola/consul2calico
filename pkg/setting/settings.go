package setting

import (
	capi "github.com/hashicorp/consul/api"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

//global settings

var ServiceMap      = make(map[string][]*capi.CatalogService)
var ConsulCalicoMap = make(map[string]string)
var MapAction 	    = make(map[string]map[string]time.Time)

func UpdateGlobalParams(path string)  {

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
		ConsulCalicoMap[k]=v
		MapAction[k]= make(map[string]time.Time)
	}


}