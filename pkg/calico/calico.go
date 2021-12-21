package calico

import (
	"context"
	"github.com/cenkalti/backoff"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	client3 "github.com/projectcalico/libcalico-go/lib/clientv3"
	"github.com/projectcalico/libcalico-go/lib/options"
	"github.com/taboola/consul2calico/pkg/log"
	"github.com/taboola/consul2calico/pkg/utils"
	"os/signal"
	"syscall"

	"os"
)

var (
	logger = log.LoggerFunction
)

func init() {
	log.SeValues("calico-controller", "", "")
}

type Controller interface {
	GetNetworkSet(context.Context, string) (*v3.GlobalNetworkSet, error)
	UpdateNetworkSet(ctx context.Context, netName string, ipsAdd []string,
		ipsDelete []string, stop chan os.Signal) error
}

type ControllerImpl struct {
	calicoClient client3.Interface
}

func New(c client3.Interface) *ControllerImpl {
	return &ControllerImpl{c}
}

func (c *ControllerImpl) GetNetworkSet(ctx context.Context, netName string) (*v3.GlobalNetworkSet, error) {
	var net *v3.GlobalNetworkSet
	var Err error
	logger().Debugf("Trying to GET GlobalNetworkSet: %v", netName)
	errAfterRetry := backoff.Retry(func() error {
		net, Err = c.calicoClient.GlobalNetworkSets().Get(ctx, netName, options.GetOptions{})
		if Err != nil {
			logger().Errorf("Failed to GET GlobalNetworkSet (%v) info from Calico ERROR :%v",
				netName, Err)
			return Err
		}
		return nil
	}, utils.GetBackOff())

	if errAfterRetry != nil {
		logger().Errorf("Failed to GET GlobalNetworkSet (%v) info from Calico for few times ERROR :%v",
			netName, errAfterRetry)
		return nil, errAfterRetry
	}
	logger().Debugf("Success to GET GlobalNetworkSet (%v) info from Calico", netName)
	return net, nil
}

func (c *ControllerImpl) UpdateNetworkSet(ctx context.Context, netName string, ipsAdd []string,
	ipsDelete []string, stop chan os.Signal) error {
	logger().Debugf("Updating %v Ips to GlobalNetworkSet: %v", len(ipsAdd)+len(ipsDelete), netName)
	//Get globalNetworkSet.
	net, Err := c.GetNetworkSet(ctx, netName)
	if Err != nil {
		logger().Errorf("Failed - trying to get GlobalNetworkSet: %v ERROR : %v", netName, Err)
		signal.Notify(stop, syscall.SIGTERM)
	}
	//Servers to add
	if len(ipsAdd) > 0 {
		net.Spec.Nets = append(net.Spec.Nets, ipsAdd...)
	}
	//Servers to delete
	if len(ipsDelete) > 0 {
		l, _ := utils.CompareSlice(net.Spec.Nets, ipsDelete)
		net.Spec.Nets = l
	}

	errAfterRetry := backoff.Retry(func() error {
		_, err := c.calicoClient.GlobalNetworkSets().Update(ctx, net, options.SetOptions{TTL: 0})
		if err != nil {
			logger().Errorf("Failed - trying to UPDATE GlobalNetworkSet(%v) ERROR : %v",
				net, err)
			return err
		}
		return nil
	}, utils.GetBackOff())

	if errAfterRetry != nil {
		logger().Errorf("Failed - trying to UPDATE GlobalNetworkSet(%v) for few times ERROR :%v",
			net, errAfterRetry)
		return errAfterRetry
	}
	logger().Debugf("Success - trying to UPDATE GlobalNetworkSet: %v", net)
	return nil

}

// GetCalicoClient Returns calico client initialized from ENV
func GetCalicoClient() (client3.Interface, error) {
	var client client3.Interface
	var calicoErr error
	logger().Info("Starting to initialize calico client")
	errAfterRetry := backoff.Retry(func() error {
		client, calicoErr = client3.NewFromEnv()
		if calicoErr != nil {
			logger().Error(calicoErr)
			return calicoErr
		}
		return nil
	}, utils.GetBackOff())

	if errAfterRetry != nil {
		logger().Errorf("Failed to initialize calico client For few times ERROR : %v", errAfterRetry)
		return nil, errAfterRetry
	}
	logger().Info("Success initialize calico client")
	return client, nil
}
