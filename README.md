# Consul Calico Sync

<center>

<img src="https://git.taboolasyndication.com/projects/ITP/repos/taboola.com/raw/k8s/consul-calico-sync/docs/design/high-lvl.svg?at=refs%2Fheads%2FK8S-825-consul-calico-sync-refactor-ui" width="800" height="600">
</center>

## Overview

This project will sync/configure calico network policies based on consul KV state.

It will allow ingress/egress traffic from nodes registered in consul to deployments running on kubernetes .

Whenever a node is added to the Hostgroup / rebuilt / changes ip , this project will dynamically change the corresponding calico GlobalNetworkSet.


## Getting Started Running with Helm

1. Create ETCD secret: 

    ``` bash
    kubectl create secret generic etcd-cert \
    --from-file=etcd-ca.crt=./etcd-ca.crt.txt \
    --from-file=etcd.crt=./etcd.crt.txt \
    --from-file=etcd.key=./etcd.key.txt 
    ```

2. Build docker image :
    
    ``` bash
    docker build -t consul-calico-sync:0.0.1 .
    ```

3. Push to local repository :

    ``` bash
    docker push  http://local-repo:8080/consul-calico-sync:0.0.1 .
    ```

4. Change image in values.yaml

    ``` bash
    # The name (and tag) of the Docker image for consul2calico sync.
    image:
    repository: http://local-repo:8080/consul-calico-sync
    pullPolicy: Always
    tag: 0.0.1
    ```

5. Install chart 

    ``` bash
    helm install -n consul-calico-sync -c ./charts/ --namespace namespace
    ```


## How to run tests :

Defaults configured for tests :
    ```
    CALICO_SYNC_INTERVAL=2s
    CALICO_REMOVE_GRACE_TIME=30m
    ```

- With logs :
    ``` bash
    go test  ./...
    ```

- Without logs :
    ``` bash
    go test  ./... -v
    ```


## Future releases

- Add support for  Kubernetes API datastore . (Currently this project support Calico deployments with etcd as datastore)
- Add support for consul TLS .