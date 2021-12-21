module consul-sync

go 1.16

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/consul/api v1.11.0
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/projectcalico/api v0.0.0-20210727230154-ae822ba06c23
	github.com/projectcalico/libcalico-go v1.7.2-0.20210809162050-073c4ac02f2b
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	go.etcd.io/etcd v0.5.0-alpha.5.0.20201125193152-8a03d2e9614b
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sys v0.0.0-20211110154304-99a53858aa08 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/apimachinery v0.21.0-rc.0
)

replace (
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.2
	github.com/sirupsen/logrus => github.com/projectcalico/logrus v1.0.4-calico
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.5
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
	google.golang.org/protobuf => google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/tchap/go-patricia.v2 => github.com/tchap/go-patricia/v2 v2.3.1
)
