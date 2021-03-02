# gorestapi

My personal Go playground.  

REST api written in Go and K8s deployment.  

## K8s installation

This K8s cluster is crio version from my other repo. Platform is Ubuntu 20.04 + KVM  

kubectl apply -f gorestapi-deployment.yaml  

kubectl apply -f gorestapi-svc.yaml  

cmd line api testing (first version of restapi app):  

curl -i <http://127.0.0.1:8080/ping>  

curl -i -H "Content-type: application/json" -d '{"title":"Hello","post":"World"}' <http://127.0.0.1:8080/newsfeed>  

curl -i <http://127.0.0.1:8080/newsfeed>  

K8s cmd line testing:  
kubectl get svc  

```plaintext:
NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE  
gorestapi    ClusterIP   10.110.175.227   <none>        8080/TCP   4m55s  
kubernetes   ClusterIP   10.96.0.1        <none>        443/TCP    58d
```

Check if app is responding:  
curl -i <http://10.110.175.227:8080/ping>  
HTTP/1.1 200 OK  
Content-Type: application/json; charset=utf-8  
Date: Mon, 08 Feb 2021 13:49:24 GMT  
Content-Length: 20  

{"hello":"Found me"}  

## MetalLB loadbalancer installation

kubectl get configmap kube-proxy -n kube-system -o yaml | \  
sed -e "s/strictARP: false/strictARP: true/" | \  
kubectl apply -f - -n kube-system  

kubectl apply -f <https://raw.githubusercontent.com/metallb/metallb/v0.9.5/manifests/namespace.yaml>  

kubectl apply -f <https://raw.githubusercontent.com/metallb/metallb/v0.9.5/manifests/metallb.yaml>  

kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"  

$ cat config.yaml  

```yaml:

apiVersion: v1  
  kind: ConfigMap  
  metadata:  
    namespace: metallb-system  
    name: config  
  data:  
    config: |  
      address-pools:  
      - name: restapipool
        protocol: layer2  
        addresses:  
        - 10.0.1.245-10.0.1.250  

```

Above config will give addr pool 245 - 250 to MetalLB

kubectl apply -f config.yaml  

Test IP routing: Change gorestapi-svc.yaml type: ClusterIP => type: LoadBalancer  

Check if app gets routable external ip:  
kubectl get svc  

```plaintext:
NAME         TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE  
gorestapi    LoadBalancer   10.110.175.227   10.0.1.245    8080:32370/TCP   9m19s  
kubernetes   ClusterIP      10.96.0.1        <none>        443/TCP          58d  
```

Test from outside:  

```plaintext:
  curl -i <http://10.0.1.245:8080/ping>
  HTTP/1.1 200 OK
  Content-Type: application/json; charset=utf-8
  Date: Mon, 08 Feb 2021 13:56:31 GMT
  Content-Length: 20

  {"hello":"Found me"}
```

## Install haproxy-ingress controller

More info: <https://github.com/jcmoraisjr/haproxy-ingress/tree/master/examples/deployment>

Label nodes 1-3 (I have 4 worker nodes):  
  kubectl label node worker1 role=ingress-controller
  node/worker1 labeled  
  kubectl label node worker2 role=ingress-controller
  node/worker2 labeled  
  kubectl label node worker3 role=ingress-controller
  node/worker3 labeled  

Check labels:
kubectl get nodes --selector='role=ingress-controller'  

```plaintext:
NAME      STATUS   ROLES    AGE   VERSION  
worker1   Ready    <none>   57d   v1.20.0  
worker2   Ready    <none>   57d   v1.20.0  
worker3   Ready    <none>   57d   v1.20.0  
```

kubectl create ns ingress-controller  

openssl req \  
  -x509 -newkey rsa:2048 -nodes -days 365 \  
  -keyout tls.key -out tls.crt -subj '/CN=localhost'  

kubectl --namespace ingress-controller create secret tls tls-secret --cert=tls.crt --key=tls.key  

rm -v tls.crt tls.key  

kubectl apply -f haproxy-ingress-deployment.yaml  

Check ingress-controller daemonset:  
kubectl get daemonsets -n ingress-controller  
NAME              DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR             AGE  
haproxy-ingress   3         3         3       3            3           role=ingress-controller   70s

kubectl apply -f haproxy-svc.yaml  

kubectl apply -f gorestapi-ingress.yaml  

Now we have MetalLB to give static IP 10.0.1.248 to Haproxy which is Watching for ingress class 'haproxy' to route traffic to restapi Pod port 8080.  

Test from outside of K8s:  

curl -i <http://10.0.1.248:8080/ping>  
HTTP/1.1 200 OK  
content-type: application/json; charset=utf-8  
date: Tue, 09 Feb 2021 13:05:32 GMT  
content-length: 20  
strict-transport-security: max-age=15768000  

{"hello":"Found me"}  

## Postgres DB installation

Note! This is only for my playground, I'm not using persistent volumes here.  

kubectl apply -f postgres-configmap.yaml  

kubectl apply -f postgres-deployment-svc.yaml  

Note! We have type: NodePort to access our db from host.  

Check Postgres port:  

kubectl get svc  
postgres     NodePort    10.100.109.205   none        5432:32531/TCP   28m  

Test db connection: https:  

psql -h 10.0.1.204 -U admin --password -p 32531 omadb  
Password:
omadb=# quit  

At this stage, I developed second version of go app with Postgres db connection.

## Wait for postgres connection with initContainer

Get script:  
git clone <https://github.com/bells17/wait-for-it-for-busybox>  

Create configmap:  
kubectl create configmap wait-for-it --dry-run=client -o yaml --from-file=wait-for-it.sh > wait-for-it-configmap.yaml  

Deploy configmap (I always like to checkit up before deployment):  
kubectl apply -f wait-for-it-configmap.yaml  
configmap/wait-for-it created  

Testing above wait-for-it script with standalone alpine, I just didn't get it work in Busybox:  
kubectl logs alpine  
wait-for-it.sh: waiting for postgres:5432 without a timeout  
wait-for-it.sh: postgres:5432 is available after 0 seconds  

Same logs from initContainer:  
kubectl logs gorestapi-566b5db78b-z87qt -c wait-for-postgres  
wait-for-it.sh: waiting for postgres:5432 without a timeout  
wait-for-it.sh: postgres:5432 is available after 0 seconds  

## Special notes

In Dockerfile:  
RUN go get github.com/golang-migrate/migrate/v4/database/postgres  
RUN go get github.com/golang-migrate/migrate/v4/source/file  

Above is needed for not having missing import error

In gorestapi-ingress.yaml:  
ingress.kubernetes.io/ssl-redirect: "false"  

Above is needed because default haproxy config for ssl redirect is 'true' and this demo app does not have ssl.  

haproxy-ingress-deployment.yaml RBAC fix:  

```yaml:
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
  - apiGroups:
      - "extensions"
      - "networking.k8s.io"
    resources:
      - ingresses
      - ingressclasses
      - ingresses/status
```

Rsyslog sidecar for http access logs in haproxy-ingress-deployment.yaml:  

```yaml:
      - name: access-logs
        image: jumanjiman/rsyslog
        ports: 
        - name: udp
          containerPort: 514
          protocol: UDP
```

## Haproxy connection logging

http-log-format: "%ci:%cp\\ method=%HM\\ uri=%HU\\ rcvms=%TR\\ serverms=%Tr\\ activems=%Ta\\ bytes=%B\\ status=%ST"  
%ci = client_ip  
%cp = client_port  
%HM = http method (GET / POST)  
%HU = request URI path  
%TR = request timeout  
%Tr = response time  
%Ta = active time of the request  
%B  = bytes_read  
%ST = status_code  

In this demo I had two different sidecar alternatives for collecting haproxy logs: rsyslog and netcat. Here's rsyslog example:  
kubectl logs -f haproxy-ingress-bnn7s -n ingress-controller -c access-logs  

```plaintext:
2021-02-24T13:18:33.493980+00:00 localhost 10.0.1.1: 44810 method=GET uri=/data rcvms=0 serverms=2 activems=2 bytes=87 status=200
2021-02-24T13:18:34.516629+00:00 localhost 10.0.1.1: 44812 method=GET uri=/data rcvms=0 serverms=3 activems=4 bytes=87 status=200
```

## Installing Prometheus with Helm 3

I don't have storage in my demo lab, so I need to disable peristentVolumeClaims:

```plaintext:
helm install prometheus prometheus-community/prometheus -n monitoring --set alertmanager.persistentVolume.enabled=false --set server.persistentVolume.enabled=false --set pushgateway.persistentVolume.enabled=false
```

Check it up: kubectl get pods -n monitoring

```plaintext:
NAME                                            READY   STATUS    RESTARTS   AGE
prometheus-alertmanager-6866b96d6f-fk2qt        2/2     Running   0          89s
prometheus-kube-state-metrics-95d956569-f7gx6   1/1     Running   0          89s
prometheus-node-exporter-2vscs                  1/1     Running   0          89s
prometheus-node-exporter-b27hw                  1/1     Running   0          89s
prometheus-node-exporter-hcgzd                  1/1     Running   0          89s
prometheus-node-exporter-kn4ps                  1/1     Running   0          89s
prometheus-pushgateway-bd8d484d6-d5wcr          1/1     Running   0          89s
prometheus-server-7f67fc9bdb-2mhqx              2/2     Running   0          89s
```

Default K8s installation need some fixing for Prometheus:  

```plaintext:
$ kubectl get pod kube-controller-manager-master1 -n kube-system  -o yaml | grep 'bind-address'
    - --bind-address=127.0.0.1
```

We have some problems to fix:

1. Prometheus is trying to get kube controller manager metrics from deprecated port 10252, new port is 10259  
2. Same as above with kube scheduler, prome is trying deprecated port 10251, new port is 10257  
3. Prometheus is trying to access master ip, when as seen above, bind-adderess is 127.0.0.1  

I'll use haproxy to expose above metrics, my custom build jrcjoro1/haproxy-fix:1.8 will redirect prometheus scrape to localhost and correct port. For testing haproxy-fix redirect, I used curl-test.yaml:

kubectl apply -f haproxy-fix-deployment.yaml  
kubectl apply -f curl-test.yaml

## Install Haproxy monitoring support

First remove old installation:  
kubectl delete -f haproxy-ingress-deployment.yaml  
kubectl delete -f haproxy-svc.yaml  
kubectl delete ns ingress-controller  

Open connection to Prometheus server:  
kubectl --address localhost,10.0.1.131 -n monitoring port-forward prometheus-server-7f67fc9bdb-2mhqx 8090:9090

Test from outside of cluster (that query will list all K8s resources):  

```plaintext:
curl --data-urlencode 'query=up{}' http://10.0.1.131:8090/api/v1/query | jq  
```

Above Prometheus version has problem with kind ServiceMonitor, so I'll try next version:  

```plaintext:
helm install prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring --set prometheusOperator.hostNetwork=true --set defaultRules.rules.kubernetesStorage=false --set prometheusOperator.tls.internalPort=10251
```

Test again from outside:  

```plaintext:
kubectl --address localhost,10.0.1.131 -n monitoring port-forward svc/prometheus-stack-kube-prom-prometheus 32090:9090  
```

```plaintext:
curl --data-urlencode 'query=up{}' http://10.0.1.131:32090/api/v1/query | jq
```

## Re-install haproxy with helm and monitoring

helm search repo haproxy-ingress

```plaintext:
NAME                            CHART VERSION APP VERSION DESCRIPTION  
haproxy-ingress/haproxy-ingress 0.12.0        v0.12       Ingress controller for HAProxy loadbalancer  
```

First let's check default values:  
helm pull haproxy-ingress/haproxy-ingress  

```plaintext:
helm install haproxy-ingress haproxy-ingress/haproxy-ingress --create-namespace --namespace ingress-controller --version 0.12.0 --set controller.hostNetwork=true --set controller.stats.enabled=true --set controller.metrics.enabled=true --set controller.serviceMonitor.enabled=true
```

First test haproxy metrics api:  

```plaintext:
curl -i http://10.104.136.22:9101/metrics
```

Install Haproxy configmap:  
kubectl apply -f haproxy-configmap.yaml  
