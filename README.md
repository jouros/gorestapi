# gorestapi

My personal Go playground.  

REST api written in Go and K8s deployment: MetalLB loadbalancer, Prometheus monitoring, Haproxy, Postgres, Helm 3, Lens, Krew, RBAC, Falco, Secrets management, GitOps and plenty of admin tricks.

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

Note! Here is first tested Prometheus helm version, I changed version later.

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

```plaintext:
$ kubectl describe pod etcd-master1 -n kube-system | grep url
Annotations:          kubeadm.kubernetes.io/etcd.advertise-client-urls: https://10.0.1.131:2379
      --advertise-client-urls=https://10.0.1.131:2379
      --initial-advertise-peer-urls=https://10.0.1.131:2380
      --listen-client-urls=https://127.0.0.1:2379,https://10.0.1.131:2379
      --listen-metrics-urls=http://127.0.0.1:2381
      --listen-peer-urls=https://10.0.1.131:2380
```

etcd configs can be found from here: /etc/kubernetes/manifests/etcd.yaml  

Open connection to Prometheus server:  
kubectl --address localhost,10.0.1.131 -n monitoring port-forward prometheus-server-7f67fc9bdb-2mhqx 8090:9090

Test from outside of cluster (that query will list all K8s resources):  

curl --data-urlencode 'query=up{}' <http://10.0.1.131:8090/api/v1/query> | jq  

Above prometheus-community/prometheus version has problem with kind ServiceMonitor, so I'll try next version kube-prometheus-stack:  

```plaintext:
helm install prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring --set prometheusOperator.hostNetwork=true --set defaultRules.rules.kubernetesStorage=false --set prometheusOperator.tls.internalPort=10251
```

Open Prometheus connection:  

```plaintext:
kubectl --address localhost,10.0.1.131 -n monitoring port-forward svc/prometheus-stack-kube-prom-prometheus 32090:9090  
```

Test again from outside:  

curl --data-urlencode 'query=up{}' <http://10.0.1.131:32090/api/v1/query> | jq  

We have some access problems to be fixed:

1. Prometheus is trying to get kube controller manager metrics from deprecated port 10252, new port is 10259  
2. Same as above with kube scheduler, prome is trying deprecated port 10251, new port is 10257  
3. Prometheus is trying to access master ip, when as seen above, bind-adderess is 127.0.0.1
4. Prometheus is trying to access etcd from <http://masterip:2379>, but as we can see from above, etcd is offering https connection
5. Prometheus is trying to access kube-proxy metrics from masterip:10249 when it's available in localhost:10249

I'll use haproxy to fix above problems 1-3. I made custom build jrcjoro1/haproxy-fix that will redirect prometheus scrape to correct Pod localhost.

For testing haproxy-fix redirect, I used curl-test.yaml:

kubectl apply -f haproxy-fix-deployment.yaml  
kubectl apply -f curl-test.yaml

For problem 4. we need to set client cert auth. First let's check etcd connection from control-plane cmd line:  

```plaintext:
sudo curl --cert /etc/kubernetes/pki/etcd/peer.crt --key /etc/kubernetes/pki/etcd/peer.key --cacert /etc/kubernetes/pki/etcd/ca.crt  https://10.0.1.131:2379/health
{"health":"true"}
```

Create Prometheus secret etcd-client:  
./generate_prome_etcd_auth.sh  
secret/etcd-client created  
NAME          TYPE     DATA   AGE  
etcd-client   Opaque   3      0s  

Re-install prometheus-stack with etcd auth secret:  

helm uninstall prometheus-stack -n monitoring  
release "prometheus-stack" uninstalled  

Complete wipeout:
kubectl delete ns monitoring  

Re-install etcd secrets:  
./generate_prome_etcd_auth.sh  
secret/etcd-client created  
NAME          TYPE     DATA   AGE  
etcd-client   Opaque   3      1s  

Install Prometheus with custom values:  
helm install -f custom-prometheus-values.yaml prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring  

Fixing problem 5. First check kube-proxy metrics api:  
curl -i <http://127.0.0.1:10249/metrics>  

Edit kube-proxy configmap:  
kubectl edit cm kube-proxy -n kube-system <= change metricsBindAddress: "" => metricsBindAddress: 0.0.0.0

Reload kube-proxy:  
kubectl delete pod -l k8s-app=kube-proxy -n kube-system  

Re-test with control-plane ip:  
curl -i <http://10.0.1.131:10249/metrics>  

Now Prometheus can discover all targets :)

## Re-install haproxy with helm and monitoring

First remove old non Helm installation:  
kubectl delete -f haproxy-ingress-deployment.yaml  
kubectl delete -f haproxy-svc.yaml  
kubectl delete ns ingress-controller  

helm search repo haproxy-ingress

```plaintext:
NAME                            CHART VERSION APP VERSION DESCRIPTION  
haproxy-ingress/haproxy-ingress 0.12.0        v0.12       Ingress controller for HAProxy loadbalancer  
```

First let's check default values:  
helm pull haproxy-ingress/haproxy-ingress  

```plaintext:
helm install haproxy-ingress haproxy-ingress/haproxy-ingress --create-namespace --namespace ingress-controller --version 0.12.0 --set controller.hostNetwork=true --set controller.stats.enabled=true --set controller.metrics.enabled=true --set controller.serviceMonitor.enabled=true --set-string controller.metrics.service.annotations."prometheus\.io/port"="9101" --set-string controller.metrics.service.annotations."prometheus\.io/scrape"="true"
```

Test haproxy metrics api:  

curl -i <http://10.104.136.22:9101/metrics>  

Install Haproxy configmap:  
kubectl apply -f haproxy-configmap.yaml  

Check servicemonitor:  
kubectl get servicemonitors -n ingress-controller  
NAME              AGE  
haproxy-ingress   87s  

Check haproxy annotations:  
kubectl describe service haproxy-ingress-metrics -n ingress-controller | grep prometheus  
prometheus.io/port: 9101  
prometheus.io/scrape: true  

Update prometheus with haproxy scape configs:  

```plaintext:
helm upgrade --reuse-values -f custom-prometheus-values2.yaml prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring
Release "prometheus-stack" has been upgraded. Happy Helming!
```

--reuse-values will merge additional custom values to chart and keep previous settings in placce
--reset-values would reset all values back to default values.yaml chart except those provided by custom chart  

Test if you can see haproxy-ingress and haproxy-exporter up:
curl --data-urlencode 'query=up{}' <http://10.0.1.131:32090/api/v1/query> | jq  

Alternatively you can point browser to <http://http://10.0.1.131:32090/targets> and check if above targets are grean.  

If ok, you can now use all haproxy_ related functions to query metrics.  

## Lens K8s monitoring

Download Lens: <https://github.com/lensapp/lens/releases/tag/v4.1.4>
Install: sudo apt install ./Lens-4.1.4.amd64.deb

Copy ~/.kube/config of the remote Kubernetes host to your local dir.  

Add cluster to Lens by giving path to above config.  

## Install krew

Krew is kubectl pluging to find and get other kubectl plugins.  

Install info: <https://krew.sigs.k8s.io/docs/user-guide/setup/install/>

curl -fsSLO "<https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.tar.gz>"

tar zxvf krew.tar.gz  

./krew-linux_amd64 install krew

## Kubernetes RBAC

ServiceAccount: Communication with kube api server if you do not specify a service account, it is automatically assigned the default service account:  
kubectl get pod testcurl -o yaml | grep 'serviceAccount'  
  serviceAccount: default  
  serviceAccountName: default  

Default namespace does not have anyt other ServiceAccounts:  
kubectl get serviceaccounts -n default  
NAME      SECRETS   AGE  
default   1         86d  

While in ingress-controller namespace we have:  
kubectl get serviceaccounts -n ingress-controller  
NAME              SECRETS   AGE  
default           1         3d1h  
haproxy-ingress   1         3d1h  

Role: Set permissions within namespace:  
  apiGroups: kubectl api-resources -o wide  
  resources: target objects, eg. configmaps, pods, secrets, namespaces  
  verbs: what you can do with ohject, eg. get, create, update, watch, list, patch, delete, deletecollection  

Cluster role: Set cluster wide permissions.  

Role binding: Link ServiceAccount and Role together: roleRef (kind, name, apiGroup) + subjects (entity that will make operations)

Cluster role binding: Grant cluster wide access: roleRef + subjects

Default roles in Kubernetes are: view, edit, admin, cluster-admin

For example:
kubectl get serviceAccounts -n ingress-controller  
NAME              SECRETS   AGE  
default           1         3d23h  
haproxy-ingress   1         3d23h  

We have haproxy-ingress serviceAccount in ingress-controller namespace

Let's install rbac-lookup: kubectl krew install rbac-lookup  

haproxy-ingress has ClusterRole haproxy-ingress:  

kubectl-rbac_lookup haproxy-ingress  
SUBJECT            SCOPE                ROLE  
haproxy-ingress    ingress-controller   Role/haproxy-ingress  
haproxy-ingress    ingress-controller   Role/haproxy-ingress  
haproxy-ingress    cluster-wide         ClusterRole/haproxy-ingress  
haproxy-ingress    cluster-wide         ClusterRole/haproxy-ingress  

Other userful commands:
kubectl-rbac_lookup --kind user  
kubectl-rbac_lookup --kind group

haproxy-ingress ClusterRole is granted to:

```plaintext:
kubectl describe clusterrole haproxy-ingress
Name:         haproxy-ingress
Labels:       app.kubernetes.io/instance=haproxy-ingress
              app.kubernetes.io/managed-by=Helm
              app.kubernetes.io/name=haproxy-ingress
              app.kubernetes.io/version=v0.12
              helm.sh/chart=haproxy-ingress-0.12.0
Annotations:  meta.helm.sh/release-name: haproxy-ingress
              meta.helm.sh/release-namespace: ingress-controller
PolicyRule:
  Resources                           Non-Resource URLs  Resource Names  Verbs
  ---------                           -----------------  --------------  -----
  events                              []                 []              [create patch]
  services                            []                 []              [get list watch]
  ingressclasses.extensions           []                 []              [get list watch]
  ingresses.extensions                []                 []              [get list watch]
  ingressclasses.networking.k8s.io    []                 []              [get list watch]
  ingresses.networking.k8s.io         []                 []              [get list watch]
  nodes                               []                 []              [list watch get]
  configmaps                          []                 []              [list watch]
  endpoints                           []                 []              [list watch]
  pods                                []                 []              [list watch]
  secrets                             []                 []              [list watch]
  ingresses.extensions/status         []                 []              [update]
  ingresses.networking.k8s.io/status  []                 []              [update]
```

So from above we can see that user haproxy-ingress can list and watch pods, you can verify same info with can-i:  
kubectl auth can-i list pods --as haproxy-ingress  
yes  

Can user haproxy-ingress delete pods?
kubectl auth can-i delete pods --as haproxy-ingress
no

With rakkess we can inspect authorizations granted to user:  
kubectl krew install access-matrix

To get full list of grants for serviceAccount haproxy-ingress: kubectl access-matrix -n ingress-controller --as haproxy-ingress  

With kubectl-who-can we can find subjects that can perform a spesific action:  
kubectl krew install who-can

For example subjects that can list pods in ingress-controller namespace:
kubectl who-can list pods -n ingress-controller

## Kubernetes Auditing

In Kubernetes auditing first we need to define audit policy that will define rules what will be recorded and what data will be included. In audit-policy.yaml defined audit levels are:  
none: don't log events  
Metadata: log request metadata  
Request:  log event metadata and request body  
RequestResponse: log event metadata, request and response bodies  

Set audit backend for kube-apiserver:

Add /etc/kubernetes/maniifests/kube-apiserver.yaml:
    - kube-apiserver
    - --audit-log-path=/var/log/audit.log
    - --audit-log-maxage=5
    - --audit-log-maxbackup=5
    - --audit-log-maxsize=1
    - --audit-log-truncate-enabled
    - --audit-policy-file=/etc/kubernetes/audit-policy.yaml

Set mount options for audit:

```plaintext:
    volumeMounts:
    - mountPath: /etc/kubernetes/audit-policy.yaml
      name: audit
      readOnly: true
    - mountPath: /var/log/audit.log
      name: audit-log
      readOnly: false

  volumes: 
  - hostPath:
      path: /etc/kubernetes/audit-policy.yaml 
      type: File
    name: audit
  - hostPath:
      path: /var/log/audit.log
      type: FileOrCreate
    name: audit-log
```

kube-apiserver will watch for config changes and reload automatically.

## Kubernetes disaster recovery, how to re-install cluster

kubeadm reset: clean up files that were created by kubeadm init or join. When executed in control-plane node, wipes out all info from previous cluster and print out join info to new cluster. You have to re-join all worker nodes by executing kudeadm reset + kubeadm join printed out.

In my setup I did reset on all nodes, kubeadm init on control-plane and join on worker nodes. I also had to execute below in every node:  
cat <<EOF | sudo tee /etc/modules-load.d/crio.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

All labels are wiped out, so I had to re-label.  

I got this error: "failed to set bridge addr: "cni0" already has an IP address" for some starting Pod. I checked which node Pod was running and executed in that node:  
sudo ip link set cni0 down  
sudo brctl delbr cni0

## Deploy Falco security

helm repo add falcosecurity <https://falcosecurity.github.io/charts>  
"falcosecurity" has been added to your repositories  

helm search repo falco  
NAME                         CHART VERSION APP VERSION DESCRIPTION  
falcosecurity/falco          1.7.7         0.27.0      Falco  
falcosecurity/falco-exporter 0.5.1         0.5.0       Prometheus Metrics Exporter for Falco output ev...  
falcosecurity/falcosidekick  0.2.9        2.21.0     A simple daemon to help you with falco's outputs  

Check Falco default values:  
helm pull falcosecurity/falco

To be continued...

## Manage Kubernetes secrets

To be continued...  

## Kubernetes GitOps

To be continued...
