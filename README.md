# geolocation-go [![Docker](https://github.com/adavarski/REST-go-k8s-geolocation/actions/workflows/docker.yml/badge.svg)](https://github.com/adavarski/REST-go-k8s-geolocation/actions/workflows/docker.yml) [![Go](https://github.com/adavarski/REST-go-k8s-geolocation/workflows/go.yml/badge.svg)](https://github.com/adavarski/REST-go-k8s-geolocation/actions/workflows/go.yml) [![k8s](https://github.com/adavarski/REST-go-k8s-geolocation//actions/workflows/k8s.yml/badge.svg)](https://github.com/adavarski/REST-go-k8s-geolocation/actions/workflows/k8s.yml)

This repository contains a simple geolocation api microservice, fast, reliable, Kubernetes friendly and ready written in go as a proof of concept.

## Motivations

Study the feasibility of having a geolocation REST API microservice running alongside our other microservices in Kubernetes to avoid relying on the `Cloudfront-Viewer-Country` http headers.
The requirements are:

* Reliability
* Fast
* Concurrent

## Design

`geolocation-go` is written in Go, a fery fast and performant garbaged collected and concurrent programming language.

It expose a simple `GET /rest/v1/{ip}` REST endpoint.

Parameter: 

* `/rest/v1/{ip}` (string) - IPv4 

Response:

* `{"ip":"88.74.7.1","country_code":"DE","country_name":"Germany","city":"Düsseldorf","latitude":51.2217,"longitude":6.77616}`

To retrieve the country code and country name of the given IP address, `geolocation-go` use the [ip-api.com](https://ip-api.com/) real-time Geolocation API, and then cache it in-memory and in Redis for later fast retrievals.

### Flow

```
                                                                                          
                                                                                (2)
                                                                         +--------------> In-memory cache lookup
                                                                         |                       ^ 
                                                                         |                       *
                                              +------------------------+ |                       *
+-------------+            (1)                |                        | |                       * Update in-memory cache
|             |   GET /rest/v1/{ip}           |                        | |                       *
|             +------------------------------>|                        | |      (3)              *
|   Client    |                               |     geolocation-go     | +--------------> Redis lookup (optional)
|             |          (5)                  |                        | |                       ^ 
|             |<------------------------------+                        | |                       *
+-------------+       200 - OK                |                        | |                       * Update Redis cache
                                              +------------------------+ |                       *
                                                                         |                       *
                                                                         |      (4)              *
                                                                         +--------------> http://ip-api.com/json/{ip} lookup (optional)
```

1) Client make an HTTP request to `/rest/v1/{ip}`

2) `geolocation-go` will lookup for in his in-memory datastore and send the response if cache HIT. In case of cache MISS, go to step 3)

3) `geolocation-go` will lookup in Redis, send the response if cache HIT and add the response in his in-memory datastore asynchronously. In case of cache MISS, go to step 4)

4) `geolocation-go` will make an HTTP call to the [ip-api.com](https://ip-api.com/docs/api:json) API, send back the response to the client and add the response to Redis and the in-memory datastore asynchronously.

## Configuration

`geolocation-go` is a 12-factor app using [Viper](https://github.com/spf13/viper) as a configuration manager. It can read configuration from environment variables or from .env files.

### Available variables

* `APP_ADDR`(default value: `:8080`). Define the TCP address for the server to listen on, in the form "host:port".

* `APP_CONFIG_NAME` (default value: `.env`). Name of the configuration file to read from.

* `APP_CONFIG_PATH` (default value: `.`). Directory containing the configuration file to read from.

* `SERVER_READ_TIMEOUT` (default value: `30s`). Maximum duration for reading the entire request, including the body (`ReadTimeout`).

* `SERVER_READ_HEADER_TIMEOUT` (default value: `10s`). Amount of time allowed to read request headers (`ReadHeaderTimeout`).

* `SERVER_WRITE_TIMEOUT` (default value: `30s`). Maximum duration before timing out writes of the response (`WriteTimeout`).

* `LOGGER_LOG_LEVEL` (default value: `info`). Logger log level. Available values are  "trace", "debug", "info", "warn", "error", "fatal", "panic" [ref](https://pkg.go.dev/github.com/rs/zerolog@v1.26.1#pkg-variables)

* `LOGGER_DURATION_FIELD_UNIT` (default value: `ms`). Set the logger unit for `time.Duration` type fields. Available values are "ms", "millisecond", "s", "second".

* `LOGGER_FORMAT` (default value: `json`). Set the logger format. Available values are "json", "console".

* `PROMETHEUS` (default value: `true`). Enable publishing Prometheus metrics.

* `PROMETHEUS_PATH` (default value: `/metrics`). Metrics handler path.

* `REDIS_CONNECTION_STRING` (default value `redis://localhost:6379`). Connection string to connect to Redis. The format is the following: `"redis://<user>:<pass>@<host>:<port>/<db>"`.

* `REDIS_KEY_TTL` (default `24h`). TTL of a redis key: Time before the key saved in redis will expire.

* `GEOLOCATION_API` (default value `ip-api`). Define which geolocation API to use to retrieve geo IP information. Available options are:

       * [`ip-api`](https://ip-api.com/)

       * [`ipbase`](https://ipbase.com/)

* `IP_API_BASE_URL` (default value: `http://ip-api.com/json/`). Base URL for the [`ip-api`](https://ip-api.com/) API. Note that https is not available with the free plan.

* `HTTP_CLIENT_TIMEOUT` (default value: `15s`). Timeout value for the http client.

* `PPROF` (default value: `false`). Enable the pprof server. When enable, `pprof` is available at `http://127.0.0.1:6060/debug/pprof`

## Go 
```
### Install Go
$ wget https://go.dev/dl/go1.18.4.linux-amd64.tar.gz
$ sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.4.linux-amd64.tar.gz

$ grep GO ~/.bashrc 
export GOROOT=/usr/local/go
export PATH=${GOROOT}/bin:${PATH}
export GOPATH=$HOME/go
export PATH=${GOPATH}/bin:${PATH}

$ go version
go version go1.18.4 linux/amd64

$ go run main.go 
{"level":"info","svc":"geolocation-go","time":"2022-08-01T18:54:40+03:00","message":"Starting server on address :8080 ..."}
{"level":"info","remote_client":"127.0.0.1:39404","user_agent":"curl/7.58.0","req_id":"cbjves29bh8hacek8s60","method":"GET","url":"/rest/v1/46.238.32.247","status":200,"size":122,"duration":164.016702,"time":"2022-08-01T18:55:28+03:00"}
{"level":"error","svc":"geolocation-go","req_id":"cbjves29bh8hacek8s60","time":"2022-08-01T18:55:28+03:00","message":"fail to cache in redis database: error: cannot save value in redis: dial tcp 127.0.0.1:6379: connect: connection refused"}


$ curl -X GET http://localhost:8080/rest/v1/46.238.32.247
{"ip":"46.238.32.247","country_code":"BG","country_name":"Bulgaria","city":"Sofia","latitude":42.7182,"longitude":23.2974}
```
## Docker deploy
```
$ docker build -t davarski/geolocation-go .
$ docker login
$ docker push davarski/geolocation-go
$ docker run -d -p 8080:8080 davarski/geolocation-go 
$ curl -X GET http://localhost:8080/rest/v1/8.8.8.8
{"ip":"8.8.8.8","country_code":"US","country_name":"United States","city":"Ashburn","latitude":39.03,"longitude":-77.5}
$ curl -X GET http://localhost:8080/rest/v1/46.238.32.247
{"ip":"46.238.32.247","country_code":"BG","country_name":"Bulgaria","city":"Sofia","latitude":42.7182,"longitude":23.2974}
```

## k8s Deploy

```
## KIND install

$ curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind

## Create cluster (CNI=Calico, Enable ingress)

$ cd kubernetes
$ kind create cluster --name devops --config cluster-config.yaml

$ kind get kubeconfig --name="devops" > admin.conf
$ export KUBECONFIG=./admin.conf 

$ kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
$ kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true


## Ingress Nginx (optional)
$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

## LoadBalancer (MetalLB)

$ kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
$ kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml

### Edit metallb-configmap-davar.yaml

$ docker network inspect -f '{{.IPAM.Config}}' kind
[{172.20.0.0/16  172.17.0.1 map[]} {fc00:f853:ccd:e793::/64   map[]}]

$ cat metallb-configmap-carbon.yaml 
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - 172.20.0.200-172.20.0.250
      
$ kubectl apply -f metallb-configmap.yaml       
```
### Deploy app


```
$ kubectl apply -f deploy/k8s/
deployment.apps/geolocation-go created
service/geolocation-go created
serviceaccount/geolocation-go created

$ kubectl get all 
NAME                                  READY   STATUS    RESTARTS   AGE
pod/geolocation-go-6c59b96779-z6wsg   1/1     Running   0          10s

NAME                     TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)        AGE
service/geolocation-go   LoadBalancer   10.96.161.8   172.20.0.200   80:31518/TCP   10s
service/kubernetes       ClusterIP      10.96.0.1     <none>         443/TCP        3h17m

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/geolocation-go   1/1     1            1           10s

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/geolocation-go-6c59b96779   1         1         1       10s

$ curl http://172.20.0.200/rest/v1/8.8.8.8
{"ip":"8.8.8.8","country_code":"US","country_name":"United States","city":"Ashburn","latitude":39.03,"longitude":-77.5}

$ curl http://172.20.0.200/rest/v1/46.238.32.247
{"ip":"46.238.32.247","country_code":"BG","country_name":"Bulgaria","city":"Sofia","latitude":42.7182,"longitude":23.2974}
```

## Monitoring

`geolocation-go` provides [Prometheus](https://prometheus.io/) metrics and comes with a [Grafana](https://grafana.com/docs/grafana/) dashboard located in `deploy/grafana/dashboard.json`.

### Installation with [`kube-prometheus-stack`](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

Install [`kube-prometheus-stack`](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) in your Kubernetes cluster:

```sh
# Add helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install the Prometheus operator and the kube-prometheus-stack
helm install \
       prometheus-operator \
       prometheus-community/kube-prometheus-stack \
       --create-namespace \
       --namespace monitoring

# Access the Grafana web interface through http://localhost:8080/ (default credentials: admin/prom-operator)

$ kubectl get all -n monitoring
NAME                                                          READY   STATUS    RESTARTS   AGE
pod/alertmanager-prometheus-operator-kube-p-alertmanager-0    2/2     Running   0          3m2s
pod/prometheus-operator-grafana-6cf9697844-6f5sl              3/3     Running   0          3m7s
pod/prometheus-operator-kube-p-operator-99dbdfdbd-hn5h2       1/1     Running   0          3m7s
pod/prometheus-operator-kube-state-metrics-84d5df9f46-d2nxv   1/1     Running   0          3m7s
pod/prometheus-operator-prometheus-node-exporter-9d2bd        1/1     Running   0          3m7s
pod/prometheus-operator-prometheus-node-exporter-n7sbn        1/1     Running   0          3m6s
pod/prometheus-prometheus-operator-kube-p-prometheus-0        2/2     Running   0          3m2s

NAME                                                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/alertmanager-operated                          ClusterIP   None            <none>        9093/TCP,9094/TCP,9094/UDP   3m2s
service/prometheus-operated                            ClusterIP   None            <none>        9090/TCP                     3m2s
service/prometheus-operator-grafana                    ClusterIP   10.96.38.21     <none>        80/TCP                       3m7s
service/prometheus-operator-kube-p-alertmanager        ClusterIP   10.96.85.231    <none>        9093/TCP                     3m7s
service/prometheus-operator-kube-p-operator            ClusterIP   10.96.187.218   <none>        443/TCP                      3m7s
service/prometheus-operator-kube-p-prometheus          ClusterIP   10.96.200.0     <none>        9090/TCP                     3m7s
service/prometheus-operator-kube-state-metrics         ClusterIP   10.96.121.64    <none>        8080/TCP                     3m7s
service/prometheus-operator-prometheus-node-exporter   ClusterIP   10.96.131.192   <none>        9100/TCP                     3m7s

NAME                                                          DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/prometheus-operator-prometheus-node-exporter   2         2         2       2            2           <none>          3m7s

NAME                                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/prometheus-operator-grafana              1/1     1            1           3m7s
deployment.apps/prometheus-operator-kube-p-operator      1/1     1            1           3m7s
deployment.apps/prometheus-operator-kube-state-metrics   1/1     1            1           3m7s

NAME                                                                DESIRED   CURRENT   READY   AGE
replicaset.apps/prometheus-operator-grafana-6cf9697844              1         1         1       3m7s
replicaset.apps/prometheus-operator-kube-p-operator-99dbdfdbd       1         1         1       3m7s
replicaset.apps/prometheus-operator-kube-state-metrics-84d5df9f46   1         1         1       3m7s

NAME                                                                    READY   AGE
statefulset.apps/alertmanager-prometheus-operator-kube-p-alertmanager   1/1     3m2s
statefulset.apps/prometheus-prometheus-operator-kube-p-prometheus       1/1     3m2s

$ kubectl port-forward -n monitoring svc/prometheus-operator-grafana 8081:80
```

To install the dashboard, go to "Menu" > "Import" > "Upload json file" and upload `deploy/grafana/dashboard.json`.

<details>
<summary>Click to expand</summary>
![Grafana-01-screenshot](https://raw.githubusercontent.com/adavarski/REST-go-k8s-geolocation/.docs/grafana-01.png)
</details>

## TODO

* [x] 404 and 405 custom handler

* [ ] Provide APM & tracing 

* [ ] Provide a Swagger endpoint

* [x] Support graceful shutdowns for interrupt signals (SIGTERM)
