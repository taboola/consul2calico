package calico

import (
	"context"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/taboola/consul2calico/pkg/utils"
	"os"
)

// ControllerFake fakes the calico controller
type ControllerFake struct {
	GlobalNetworkSets map[string]*v3.GlobalNetworkSet
}

// GetNetworkSet fakes action of getting a NetworkSet from calico
func (c *ControllerFake) GetNetworkSet(ctx context.Context, netName string) (*v3.GlobalNetworkSet, error) {
	return c.GlobalNetworkSets[netName], nil
}

// UpdateNetworkSet fakes action of updating a NetworkSet in calico
func (c *ControllerFake) UpdateNetworkSet(ctx context.Context, netName string, ipsAdd []string,
	ipsDelete []string, stop chan os.Signal) error {

	//Servers to add
	if len(ipsAdd) > 0 {
		c.GlobalNetworkSets[netName].Spec.Nets = append(c.GlobalNetworkSets[netName].Spec.Nets, ipsAdd...)
	}
	//Servers to delete
	if len(ipsDelete) > 0 {
		l, _ := utils.CompareSlice(c.GlobalNetworkSets[netName].Spec.Nets, ipsDelete)
		c.GlobalNetworkSets[netName].Spec.Nets = l
	}
	return nil
}
