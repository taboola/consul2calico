package controller

import (
	"context"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/taboola/consul2calico/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"testing"
	"time"
)

func TestRefreshCacheCalico(t *testing.T) {

	stop := make(chan os.Signal, 1)                                    // Create channel to receive OS signals
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the signals channel to receive SIGTERM
	c := InitFakeController()

	//Check cache before running RefreshCacheCalico
	//Expect to get value
	_, f := c.consulChange["consulService2"]
	if !f {
		t.Errorf("cache should include consulService2 when initializing NewFakeController")
	}
	//Expect to be empty
	_, f = c.consulChange["consulService1"]
	if f {
		t.Errorf("cache should not include consulService1 when initializing NewFakeController")
	}

	//start RefreshCacheCalico
	go c.RefreshCacheCalico(stop)
	//Sleep so RefreshCacheCalico can run
	time.Sleep(time.Second * 13)

	//Expect to get value
	_, f = c.consulChange["consulService1"]
	if !f {
		t.Errorf("cache should include consulService1 after running RefreshCacheCalico")
	}
	//Expect to get value
	_, f = c.consulChange["consulService2"]
	if !f {
		t.Errorf("cache should include consulService2 after running RefreshCacheCalico")
	}
	stop <- os.Interrupt

}

func TestSyncOnTrigger(t *testing.T) {

	// Create application context withCancel
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1) // Create channel to receive OS signals
	c := InitFakeController()

	//Get state of GNS before sync  globalNetworkSet1
	gotGlobalNetworkSet1, _ := c.calicoC.GetNetworkSet(ctx, "globalNetworkSet1")
	var wantGlobalNetworkSetSpec1 = v3.GlobalNetworkSetSpec{Nets: []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"}}
	var wantGlobalNetworkSet1 = &v3.GlobalNetworkSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       wantGlobalNetworkSetSpec1,
	}

	a, b := utils.CompareSlice(gotGlobalNetworkSet1.Spec.Nets, wantGlobalNetworkSet1.Spec.Nets)
	if len(a) > 0 || len(b) > 0 {
		t.Errorf("got %q , wanted %q ", gotGlobalNetworkSet1.Spec.Nets, wantGlobalNetworkSet1.Spec.Nets)
	}

	//Get state of GNS before sync  globalNetworkSet2
	gotGlobalNetworkSet2, _ := c.calicoC.GetNetworkSet(ctx, "globalNetworkSet2")
	var wantGlobalNetworkSetSpec2 = v3.GlobalNetworkSetSpec{Nets: []string{"10.2.0.1/32", "10.2.0.2/32", "10.2.0.3/32"}}
	var wantGlobalNetworkSet2 = &v3.GlobalNetworkSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       wantGlobalNetworkSetSpec2,
	}

	a, b = utils.CompareSlice(gotGlobalNetworkSet2.Spec.Nets, wantGlobalNetworkSet2.Spec.Nets)
	if len(a) > 0 || len(b) > 0 {
		t.Errorf("got %q , wanted %q ", gotGlobalNetworkSet2.Spec.Nets, wantGlobalNetworkSet2.Spec.Nets)
	}

	//Add service consulService1 to cache
	c.consulChange["consulService1"] = 1

	//start SyncOnTrigger
	go c.SyncOnTrigger(ctx, stop)

	//Sleep so Sync can run
	time.Sleep(time.Second * 13)

	//Check services got deleted from cache
	//Expect to be empty
	_, f := c.consulChange["consulService1"]
	if f {
		t.Errorf("SyncOnTrigger did not clear cahce for service consulService1")
	}
	//Expect to be empty
	_, f = c.consulChange["consulService2"]
	if f {
		t.Errorf("SyncOnTrigger did not clear cahce for service consulService2")
	}

	//Check GNS state after SyncOnTrigger run

	//Get state of GNS after sync  globalNetworkSet1
	gotGlobalNetworkSet3, _ := c.calicoC.GetNetworkSet(ctx, "globalNetworkSet1")
	var wantGlobalNetworkSetSpec3 = v3.GlobalNetworkSetSpec{Nets: []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.5/32"}}
	var wantGlobalNetworkSet3 = &v3.GlobalNetworkSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       wantGlobalNetworkSetSpec3,
	}

	a, b = utils.CompareSlice(gotGlobalNetworkSet3.Spec.Nets, wantGlobalNetworkSet3.Spec.Nets)
	if len(a) > 0 || len(b) > 0 {
		t.Errorf("got %q , wanted %q ", gotGlobalNetworkSet3.Spec.Nets, wantGlobalNetworkSet3.Spec.Nets)
	}

	//Get state of GNS after sync  globalNetworkSet2
	gotGlobalNetworkSet4, _ := c.calicoC.GetNetworkSet(ctx, "globalNetworkSet2")
	var wantGlobalNetworkSetSpec4 = v3.GlobalNetworkSetSpec{Nets: []string{"10.2.0.1/32", "10.2.0.2/32",
		"10.2.0.3/32", "10.2.0.5/32"}}
	var wantGlobalNetworkSet4 = &v3.GlobalNetworkSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       wantGlobalNetworkSetSpec4,
	}

	a, b = utils.CompareSlice(gotGlobalNetworkSet4.Spec.Nets, wantGlobalNetworkSet4.Spec.Nets)
	if len(a) > 0 || len(b) > 0 {
		t.Errorf("got %q , wanted %q ", gotGlobalNetworkSet4.Spec.Nets, wantGlobalNetworkSet4.Spec.Nets)
	}

	// Shutdown. Cancel application context will kill all attached tasks.
	cancel()
}

func TestSyncGlobalNetworkSet(t *testing.T) {

	// Create application context withCancel
	ctx, cancel := context.WithCancel(context.Background())

	stop := make(chan os.Signal, 1) // Create channel to receive OS signals
	c := InitFakeController()

	//Test consulService1
	c.SyncGlobalNetworkSet(ctx, "consulService1", "globalNetworkSet1", stop)
	gotGlobalNetworkSet, _ := c.calicoC.GetNetworkSet(ctx, "globalNetworkSet1")
	var wantGlobalNetworkSetSpec = v3.GlobalNetworkSetSpec{Nets: []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.5/32"}}
	var wantGlobalNetworkSet = &v3.GlobalNetworkSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       wantGlobalNetworkSetSpec,
	}
	a, b := utils.CompareSlice(gotGlobalNetworkSet.Spec.Nets, wantGlobalNetworkSet.Spec.Nets)
	if len(a) > 0 || len(b) > 0 {
		t.Errorf("got %q , wanted %q ", gotGlobalNetworkSet.Spec.Nets, wantGlobalNetworkSet.Spec.Nets)
	}

	//Test consulService2
	c.SyncGlobalNetworkSet(ctx, "consulService2", "globalNetworkSet2", stop)
	gotGlobalNetworkSet2, _ := c.calicoC.GetNetworkSet(ctx, "globalNetworkSet2")
	var wantGlobalNetworkSetSpec2 = v3.GlobalNetworkSetSpec{Nets: []string{"10.2.0.1/32", "10.2.0.2/32",
		"10.2.0.3/32", "10.2.0.5/32"}}
	var wantGlobalNetworkSet2 = &v3.GlobalNetworkSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       wantGlobalNetworkSetSpec2,
	}
	a, b = utils.CompareSlice(gotGlobalNetworkSet2.Spec.Nets, wantGlobalNetworkSet2.Spec.Nets)
	if len(a) > 0 || len(b) > 0 {
		t.Errorf("got %q , wanted %q ", gotGlobalNetworkSet2.Spec.Nets, wantGlobalNetworkSet2.Spec.Nets)
	}

	//Test added to MapAction
	//Expect to be empty
	if len(c.mapAction["consulService1"]) != 0 {
		t.Errorf("got %q , wanted %q ", len(c.mapAction["consulService1"]), 0)
	}
	//Expect to be empty
	_, f := c.mapAction["consulService2"]["10.2.0.5/32"]
	if f {
		t.Errorf("MapAction for service consulService2 has key 10.2.0.5/32 ")
	}
	//Expect to have value "10.2.0.3/32"
	got, f := c.mapAction["consulService2"]["10.2.0.3/32"]
	if !f {
		t.Errorf("MapAction for service consulService2 doesnt have key 10.2.0.3/32 ")
	}
	dontWant := time.Time{}.UTC()
	if got == dontWant {
		t.Errorf("time of deleting ths server 10.2.0.3/32 is not valid")
	}

	// Shutdown. Cancel application context will kill all attached tasks.
	cancel()
}

func TestMapLoopAction(t *testing.T) {
	c := InitFakeController()
	var wantAdd []string
	var wantDelete []string
	var empty []string

	//First check
	gotAdd, gotDelete := c.mapLoopAction("consulService1")
	wantDelete = append(wantDelete, "10.0.0.3/32")
	add := reflect.DeepEqual(empty, wantAdd)
	del := reflect.DeepEqual(gotDelete, wantDelete)
	if !add || !del {
		t.Errorf("got %q %q, wanted %q %q", gotAdd, gotDelete, wantAdd, wantDelete)
	}
	//Second check
	gotAdd, gotDelete = c.mapLoopAction("consulService2")
	wantAdd = append(wantAdd, "10.2.0.5/32")
	//wantDelete = append(wantDelete[:0], wantDelete[0+1:]...)
	add = reflect.DeepEqual(gotAdd, wantAdd)
	del = reflect.DeepEqual(gotDelete, empty)
	if !add || !del {
		t.Errorf("got %q %q, wanted %q %q", gotAdd, gotDelete, wantAdd, empty)
	}

}

func TestGetServicesChanged(t *testing.T) {

	c := InitFakeController()
	got := c.GetServicesChanged()
	var want = make(map[string]int)

	want["consulService2"] = 1
	a := reflect.DeepEqual(got, want)
	if !a {
		t.Errorf("got %q, wanted %q", got, want)
	}

	c.consulChange["consulService1"] = 1
	want["consulService1"] = 1
	a = reflect.DeepEqual(got, want)
	if !a {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestDeleteCache(t *testing.T) {

	c := InitFakeController()
	c.DeleteCache("consulService2")
	got := c.consulChange["consulService2"]

	want := 0
	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestAddCache(t *testing.T) {

	c := InitFakeController()
	c.AddCache("consulService1")
	got := c.consulChange["consulService1"]

	want := 1
	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestGetCatalogService(t *testing.T) {

	c := InitFakeController()
	ConsulServiceCatalog := c.GetCatalogService("consulService1")

	want := "10.0.0.1" + "10.0.0.2" + "10.0.0.5"
	var got string
	for _, node := range ConsulServiceCatalog {
		got = got + node.Address
	}
	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func InitFakeController() *SyncController {

	//Map consulSvc to calicoGlobalNetworkSet
	consulCalicoConf := make(map[string]string)
	consulCalicoConf["consulService1"] = "globalNetworkSet1"
	consulCalicoConf["consulService2"] = "globalNetworkSet2"
	c := NewFakeController(consulCalicoConf)

	//Register 3 nodes  in consul for service : consulService1
	//Register 3 nodes  in consul for service : consulService2
	consulFakeData := make(map[string][]string)
	consulFakeData["consulService1"] = []string{"10.0.0.1", "10.0.0.2", "10.0.0.5"}
	consulFakeData["consulService2"] = []string{"10.2.0.1", "10.2.0.2", "10.2.0.5"}

	InitConsulServices(c, consulFakeData)

	//Add 3 Ips to calico GlobalNetworkSet : globalNetworkSet1
	//Add 3 Ips to calico GlobalNetworkSet : globalNetworkSet2

	calicoFakeData := make(map[string][]string)
	calicoFakeData["globalNetworkSet1"] = []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"}
	calicoFakeData["globalNetworkSet2"] = []string{"10.2.0.1/32", "10.2.0.2/32", "10.2.0.3/32"}
	InitGlobalNetworkSet(c, calicoFakeData)

	//Add service consulService2 to cache
	c.consulChange["consulService2"] = 1

	//Mark '10.0.0.3/32' to be deleted with timestamp - 30 minutes .
	c.mapAction["consulService1"]["10.0.0.3/32"] = time.Now().UTC().Add(time.Duration(-35) * time.Minute)

	//Mark '10.2.0.5/32' to be added
	c.mapAction["consulService2"]["10.2.0.5/32"] = time.Time{}.UTC()

	return c

}
