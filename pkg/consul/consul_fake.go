package consul

import (
	"context"
	capi "github.com/hashicorp/consul/api"
	"os"
	"sync"
)

type ControllerFake struct{}

// Watch Gets a list of consul services and runs a sync for each consul service.
func (c *ControllerFake) Watch(ctx context.Context, mutexCache *sync.RWMutex, consulChange map[string]int,
	consulCalico map[string]string, serviceMap map[string][]*capi.CatalogService, stop chan os.Signal) {
	consulChange["service-1"] = 1
}
