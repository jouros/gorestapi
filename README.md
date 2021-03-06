# gorestapi

My personal Go and K8s playground.  

REST api written in Go and K8s deployment: MetalLB loadbalancer, Prometheus, Haproxy, Postgres, Helm 3, Lens, Krew, RBAC, Auditing, Falco, SOPS, Flux v2 and plenty of admin tricks.

## K8s installation

This K8s cluster is crio version from my other repo. Platform is Ubuntu 20.04 + KVM  

kubectl apply -f gorestapi-deployment.yaml  

kubectl apply -f gorestapi-svc.yaml  

K8s cmd line testing:  
kubectl get svc  

```plaintext:
NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE  
gorestapi    ClusterIP   10.110.175.227   <none>        8080/TCP   4m55s  
kubernetes   ClusterIP   10.96.0.1        <none>        443/TCP    58d
```

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

Check if app gets routable external ip (type: LoadBalancer):  
kubectl get svc  

```plaintext:
NAME         TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE  
gorestapi    LoadBalancer   10.110.175.227   10.0.1.245    8080:32370/TCP   9m19s  
kubernetes   ClusterIP      10.96.0.1        <none>        443/TCP          58d  
```

## Install haproxy-ingress controller non Helm version

More info: <https://github.com/jcmoraisjr/haproxy-ingress/tree/master/examples/deployment>

Label nodes 1-3 to mark them for ingress-controller (I have 4 worker nodes):  
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

## Postgres DB installation

Note! This setup is only for my playground, so I'm not using persistent volumes here.  

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

## Wait for postgres connection with initContainer

Problem: if gorestapi start before db is ready, it will fail, so we have to wait and check until Postgress is online.

Get wait-for-it script:  
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

## Special notes for Dockerfile and non Helm Haproxy version

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

In this non Helm Haproxy demo I had two different sidecar alternatives for collecting haproxy logs: rsyslog and netcat. Here's rsyslog example:  
kubectl logs -f haproxy-ingress-bnn7s -n ingress-controller -c access-logs  

```plaintext:
2021-02-24T13:18:33.493980+00:00 localhost 10.0.1.1: 44810 method=GET uri=/data rcvms=0 serverms=2 activems=2 bytes=87 status=200
2021-02-24T13:18:34.516629+00:00 localhost 10.0.1.1: 44812 method=GET uri=/data rcvms=0 serverms=3 activems=4 bytes=87 status=200
```

## Installing Prometheus with Helm 3

Note! Here is first tested Prometheus helm version prometheus-community/prometheus, I changed version couple steps later because it was missing Servicemonitor.

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

Above prometheus-community/prometheus version has problem with kind ServiceMonitor, so I'll try next version prometheus-community/kube-prometheus-stack:  

```plaintext:
helm install prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring --set prometheusOperator.hostNetwork=true --set defaultRules.rules.kubernetesStorage=false --set prometheusOperator.tls.internalPort=10251
```

Open Prometheus connection:  

```plaintext:
kubectl --address localhost,10.0.1.131 -n monitoring port-forward svc/prometheus-stack-kube-prom-prometheus 32090:9090  
```

Test again from outside:  

curl --data-urlencode 'query=up{}' <http://10.0.1.131:32090/api/v1/query> | jq  

We have some Prometheus access problems to be fixed:

1. Prometheus is trying to get kube controller manager metrics from deprecated port 10252, new port is 10259  
2. Same as above with kube scheduler, prome is trying deprecated port 10251, new port is 10257  
3. Prometheus is trying to access master ip, when as seen above, bind-adderess is 127.0.0.1
4. Prometheus is trying to access etcd from <http://masterip:2379>, but as we can see from above, etcd is offering https connection
5. Prometheus is trying to access kube-proxy metrics from masterip:10249 when it's available in localhost:10249

I'll use haproxy to fix above problems 1-3. I made custom build jrcjoro1/haproxy-fix that will redirect prometheus scrape to right place.  

For testing haproxy-fix redirect, I used curl-test.yaml:
kubectl apply -f haproxy-fix-deployment.yaml  
kubectl apply -f curl-test.yaml

For problem 4. we need to set client cert auth. First let's check etcd connection from control-plane cmd line:  

```plaintext:
sudo curl --cert /etc/kubernetes/pki/etcd/peer.crt --key /etc/kubernetes/pki/etcd/peer.key --cacert /etc/kubernetes/pki/etcd/ca.crt  https://10.0.1.131:2379/health
{"health":"true"}
```

Create Prometheus secret etcd-client with script:  
./generate_prome_etcd_auth.sh  
secret/etcd-client created  
NAME          TYPE     DATA   AGE  
etcd-client   Opaque   3      0s  

Re-install prometheus-stack with etcd auth secret:  

Uninstall:  
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

First remove non Helm installation:  
kubectl delete -f haproxy-ingress-deployment.yaml  
kubectl delete -f haproxy-svc.yaml  
kubectl delete ns ingress-controller  

Search for haproxy:  
helm search repo haproxy-ingress

```plaintext:
NAME                            CHART VERSION APP VERSION DESCRIPTION  
haproxy-ingress/haproxy-ingress 0.12.0        v0.12       Ingress controller for HAProxy loadbalancer  
```

Check default values:  
helm pull haproxy-ingress/haproxy-ingress  

Install:  

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

Some notes about helm flags:  
'reuse-values' will merge additional custom values to chart and keep previous settings in place.  
'reset-values' would reset all values back to default values.yaml chart except those provided by custom chart.  

Test if you can see haproxy-ingress and haproxy-exporter up:  
curl --data-urlencode 'query=up{}' <http://10.0.1.131:32090/api/v1/query> | jq  

Alternatively you can point browser to <http://http://10.0.1.131:32090/targets> and check if above targets are grean.  

If ok, you can now use all haproxy_ related Prometheus functions to query metrics.  

## Lens K8s monitoring

Lens is very nice graphical user interface for K8s cluster.  

Download Lens: <https://github.com/lensapp/lens/releases/tag/v4.1.4>
Install: sudo apt install ./Lens-4.1.4.amd64.deb

Copy ~/.kube/config of the remote Kubernetes host to your local dir.  

Add cluster to Lens by giving path to above config.  

## Install krew

Krew is kubectl plugin to find and get other kubectl plugins.  

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

Role binding: Link ServiceAccount and Role together: roleRef (kind, name, apiGroup) + subjects (entity that will make operations, subjects can be groups, users or ServiceAccounts)  

Cluster role binding: Grant cluster wide access: roleRef + subjects

Default roles in Kubernetes are: view, edit, admin, cluster-admin

For example:
kubectl get serviceAccounts -n ingress-controller  
NAME              SECRETS   AGE  
default           1         3d23h  
haproxy-ingress   1         3d23h  

We have haproxy-ingress serviceAccount in ingress-controller namespace

rbac-lookup: <https://github.com/FairwindsOps/rbac-lookup>  

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

Rakkess: <https://github.com/corneliusweig/rakkess>  

With rakkess we can inspect authorizations granted to user:  
kubectl krew install access-matrix

To get full list of grants for serviceAccount haproxy-ingress: kubectl access-matrix -n ingress-controller --as haproxy-ingress  

With kubectl-who-can we can find subjects that can perform a spesific action:  
kubectl krew install who-can

For example subjects that can list pods in ingress-controller namespace:
kubectl who-can list pods -n ingress-controller

## Kubernetes Auditing

In Kubernetes auditing first we need to define audit policy that will define rules of what will be recorded and what data will be included. We will use audit-policy.yaml for that, level that request is matching:  
none: don't log events  
Metadata: log request metadata  
Request:  log event metadata and request body  
RequestResponse: log event metadata, request and response bodies  

Structure of audit policy:  
**level**: none, Metadata, Request, RequestResponse  
**users**: e.g. system:serviceaccount: (singular) is the prefix for service account usernames, system:authenticated all authenticated users. system:unauthenticated all unauthenticated users etc.  
**userGroups**: e.g. system:serviceaccounts: (plural) is the prefix for service account groups  
**verbs**: get, create, update, watch, list, patch, delete  
**resources**: API groups or group + resources in that group  
**namespaces**: namespaces that this rule matches.  
**nonResourceURLs**:  Rules can apply to API resources (such as "pods" or "secrets"), non-resource URL paths (such as "/api"), or neither, but not both. If neither is specified, the rule is treated as a default for all URLs.  

For kube-apiserver testing purpose, I will use minimal-audit-policy.yaml which will log everything in metadata level. Logs will be in JSON format:  

```plaintext:
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: Metadata
```

We need to set audit backend by adding some configurations to kube-apiserver:
/etc/kubernetes/manifests/kube-apiserver.yaml:

```plaintext:
    - kube-apiserver
    - --audit-log-path=/var/log/kube-audit/audit.json
    - --audit-log-maxage=5
    - --audit-log-maxbackup=5
    - --audit-log-maxsize=1
    - --audit-log-truncate-enabled
    - --audit-policy-file=/etc/kubernetes/audit-policy.yaml
```

Set mount options for audit:

```plaintext:
    volumeMounts:
    - mountPath: /etc/kubernetes/audit-policy.yaml
      name: audit
      readOnly: true
    - mountPath: /var/log/kube-audit
      name: audit-log
      readOnly: false

  volumes: 
  - hostPath:
      path: /etc/kubernetes/audit-policy.yaml 
      type: File
    name: audit
  - hostPath:
      path: /var/log/kube-audit
      type: DirectoryOrCreate
    name: audit-log
```

Above config will create dir /var/log/kube-audit where log files will be created. You can follow files with tail -f /var/log/kube-audit/audit.json | jq  

kube-apiserver will watch for config changes and reload automatically. If you need to reload configs e.g. new audit-policy.yaml, you need to delete kube-apiserver and new Pod will be created automatically:  
kubectl delete pod/kube-apiserver-master1 -n kube-system  

If you set - --audit-log-path=- all logs will go to stdout and you can follow logs with:  
kubectl logs -f kube-apiserver-master1 -n kube-system

Repository audit-policy.yaml is just a starting point which you can use to define what events to wan't to be logged and what are just continuous high volume events.  

## Deploy Falco security

Note! I have issues with kube-apiserver, so falco deployment is still under development  

Falco is security event detection tool for kubernetes.  

helm repo add falcosecurity <https://falcosecurity.github.io/charts>  
"falcosecurity" has been added to your repositories  

helm search repo falco  
NAME                         CHART VERSION APP VERSION DESCRIPTION  
falcosecurity/falco          1.7.7         0.27.0      Falco  
falcosecurity/falco-exporter 0.5.1         0.5.0       Prometheus Metrics Exporter for Falco output ev...  
falcosecurity/falcosidekick  0.2.9        2.21.0     A simple daemon to help you with falco's outputs  

Check Falco default values:  
helm pull falcosecurity/falco

Install falco:  
helm install -f custom-falco-values.yaml falco falcosecurity/falco --create-namespace --namespace falco  

Uninstall:  
helm uninstall falco -n falco  

Falco service IP:  
kubectl get service/falco -o=jsonpath={.spec.clusterIP} -n falco

For falco we need to set below setting for /etc/kubernetes/manifests/:
--audit-webhook-config-file=/etc/kubernetes/webhook-config.yaml  

## Kubernetes GitOps with Flux v2

First I'll install hub which is a wrapper around git. It's not needed for Flux, so you can skip this step, I installed hub just because we are doing git stuff here with Flux:  
sudo apt search ^hub$  
sudo apt install hub  

git config --global hub.protocol https  
git config --list  
hub.protocol=https  

Hub will prompt for GitHub username & password the first time it needs to access the API and exchange it for an OAuth token, which it saves in ~/.config/hub  

Set couple env variables for Flux installation, I grab them from above hub installation:  
export GITHUB_TOKEN=$(cat ~/.config/hub | grep token | awk '{print $2}')  
export GITHUB_USER=$(cat ~/.config/hub | grep user | awk '{print $3}')  

Download and install flux:  
curl -s <https://toolkit.fluxcd.io/install.sh> | sudo bash  

Check version:  
flux --version  
flux version 0.9.1  

Check flux prerequisities:  
flux check --pre  

I have repo structure:  

Parameter info:  
 flux bootstrap github --help  

--owner: Github user
--repository: Repo-name. The bootstrap command creates a repository if one doesn't exist
--branch: main  
--private: false. This is public repo  
--personal: true. This is personal account: 'if true, the owner is assumed to be a GitHub user'  
--path: path in repo relative to repo root. When you have different setups like production, staging etc. or like I have test1, test2 etc. you set path to setup here

Bootstrap flux into cluster/overlays/test1:  

```plaintext:
 flux bootstrap github \
  --owner=$GITHUB_USER \
  --repository=flux-test \
  --branch=main \
  --private=false \
  --personal \
  --path=./clusters/test1
  ```

Above bootstrap will create github repo 'flux-test' with README.md and ./clusters/test1/flux-system with gotk-components.yaml, gotk-sync.yaml and kustomization.yaml and install repo deploy key.  Flux will be installed into flux-system namespace:  

A kustomization requires a source. Here the source is git repo 'flux-test', the source just points to that repo. Create a source from git repo:  

```plaintext
flux create source git podinfo \
  --url=https://github.com/jouros/flux-test \
  --branch=main \
  --interval=30s \
  --export > ./clusters/test1/test1-source.yaml
```

In above config source is checked every 30s, if the source changes, the kustomization which is related to that source, will be notified. Above config also refer to spesific branch 'main'.  

Check:  
kubectl get gitrepositories.source.toolkit.fluxcd.io -A  
or  
flux get sources git  

```plaintext:
NAMESPACE     NAME          URL                                     READY   STATUS                                                            AGE
flux-system   flux-system   ssh://git@github.com/jouros/flux-test   True    Fetched revision: main/e680931682e234b86a2808ba20ac06b3ddbf2784   30m
flux-system   podinfo       https://github.com/jouros/flux-test     True    Fetched revision: main/e680931682e234b86a2808ba20ac06b3ddbf2784   5m31s
```  

Create kustomization manifest for podinfo:  
--source: source with kind: GitRepository  
--path: path to the directory containing a kustomization.yaml file  
--prune: enable garbage collection  
--validation: client = local dry-run validation, server = APIServer dry-run  
--interval: The interval at which to retry a previously failed reconciliation  

```plaintext
flux create kustomization apps1 \
  --source=podinfo \
  --path="./apps/kustomize/test1/podinfo" \
  --prune=true \
  --validation=client \
  --interval=5m \
  --export > ./clusters/test1/apps1-kustomization.yaml
```

In above config kustomization is related to previously defined source 'podinfo'. Applications are deployded from ./apps/kustomize/test1/podinfo.  

Check:  
flux get kustomizations  

At this stage I have podinfo app running with kustomize where I changed 'minReplicas' value for cluster setup test1.  

Delete:  
flux delete kustomization apps1  
flux delete source git podinfo  

Uninstall flux:  
flux uninstall --namespace=flux-system  

Next I'll add helm GitOps:  
mkdir ./apps/base/charts/  
helm create busybox + some editing  

Test:
helm install busybox ./apps/base/charts/busybox/  

Create flux helm:  

```plaintext:
flux create hr busybox \
    --interval=10m \
    --source=GitRepository/podinfo \
    --chart=./apps/base/charts/busybox/
✚ generating HelmRelease
► applying HelmRelease
✔ HelmRelease created
◎ waiting for HelmRelease reconciliation
✔ HelmRelease busybox is ready
✔ applied revision 0.1.0
```

I'm using previously defined git source 'podinfo'.  

check:  
flux get helmreleases  

```plaintext:
NAME    READY MESSAGE                           REVISION SUSPENDED 
busybox True  Release reconciliation succeeded  0.1.0    False
```

Busybox can also be seen with:  
helm list -A  

Operations:  
I changed sleep to 600 in templates/deployment.yaml and version to 0.2.0. After git push, flux updated revision to 0.2.0.  

## Flux v2 monitoring with kube-prometheus-stack

```plaintext:
By default, Prometheus discovers PodMonitors and ServiceMonitors within its namespace, that are labeled with the same release tag as the prometheus-operator release. Sometimes, you may need to discover custom PodMonitors/ServiceMonitors, for example used to scrape data from third-party applications. An easy way of doing this, without compromising the default PodMonitors/ServiceMonitors discovery, is allowing Prometheus to discover all PodMonitors/ServiceMonitors within its namespace, without applying label filtering. To do so, you can set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues and prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues to false.
```

Update Prometheus deployment with above explained values:  

```plaintext:
helm upgrade --reuse-values prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false  
Release "prometheus-stack" has been upgraded. Happy Helming!  
```

Add gotk-monitoring.yaml to flux-test/clusters/test1/flux-system/ and push to repo. Check podMonitors:  

kubectl get podMonitors -n monitoring  

```plaintext:
NAME                      AGE  
helm-controller           49s  
kustomize-controller      49s  
notification-controller   49s  
source-controller         49s  
```

What metrics we can get:  
Ready status metrics: gotk_reconcile_condition{kind, name, namespace, type="Ready", status="True"}  
status can be: True, False, Unknown or deleted.  

e.g. number of deployments:  

```plaintext:
sum(gotk_reconcile_condition{namespace=~"default|flux-system", type="Ready", status=~"True", kind=~"Kustomization|HelmRelease"})  
```

Time spent reconciling:  

```plaintext:
gotk_reconcile_duration_seconds_bucket{kind, name, namespace, le}  
gotk_reconcile_duration_seconds_sum{kind, name, namespace}  
gotk_reconcile_duration_seconds_count{kind, name, namespace}  
```

e.g. average reconciliation:  

```plaintext:
sum(rate(gotk_reconcile_duration_seconds_sum{namespace=~"default|flux-system", kind=~"Kustomization|HelmRelease"}[5m])) by (kind)  
```

## Flux GitOps with Mozilla SOPS and Helm

My host:  
sudo dpkg --print-architecture  
amd64

Download sops:  
curl -OL <https://github.com/mozilla/sops/releases/download/v3.7.0/sops_3.7.0_amd64.deb>  

My test plan is to have sops encrypted Configmap which will overwrite default environment values from my busybox helm chart. I have data.txt values:  

```plaintext:
env:
  name1: first
  value1: "1"
  name2: second
  value2: "2"
  name3: third
  value3: "3"
```

From which I create configmap:  
kubectl create configmap busyboxdata --from-file data.txt --dry-run -o yaml > busybox-configmap.yaml  

Now I have: flux-test/apps/kustomize/test1/busybox/busybox-configmap.yaml  

```plaintext:
apiVersion: v1
data:
  data.txt: |
    env:
      name1: first
      value1: "1"
      name2: second
      value2: "2"
      name3: third
      value3: "3"
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: busyboxdata
```

I'll add this config to above dir:  

```plaintext:
cat <<EOF > ./.sops.yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    pgp: ${KEY_FP}
EOF
```

.sops.yaml:  
If your secrets are stored under a specific directory, like a git repository, you can create a .sops.yaml configuration file at the root directory to define which keys are used for which filename.  

To exclude above .sops.yaml from Flux, I add below to root of repo:  
touch .sourceignore  
echo '**/.sops.yaml' >> .sourceignore  

Encrypt:  
sops --encrypt --in-place busybox-configmap.yaml  

After above step busybox-configmap.yaml will look like this:  

```plaintext:
apiVersion: v1
data:
    data.txt: ENC[AES256_GCM,data:suNQHWl4BnerkRQx9WzoudPnq6Lul4N07A==,iv:sjhnKe51HlGXQ3keO7CSvUY+dDssPtJgt8Ug0j4fFa0=,tag:en7WIqGBKJJS6H7wHfLpug==,type:str]
kind: ConfigMap
metadata:
    creationTimestamp: null
    name: busyboxdata
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
    age: []
    lastmodified: "2021-03-25T16:36:22Z"
    mac: ENC[AES256_GCM,data:g6A0NQrL5+qAdHFySsUUG0W5P9Se7MqFaPThLhIX4u8UrAOyxxXQc0Keh/r/sVLBN/OBH3L276ywhK7o1UvK0z1zwvhdsbLsLHRDpaptAviF9ePxAbTp2U2RQJXgkxN9cbDboBUNxtl0sfu/8EZtltMVr7on2Kwnal7JTFAj1M0=,iv:A82AL4+GCQuWnYminGNeELjLW1BlXifPQ/+BqAAr5WY=,tag:I3do9vEcFwCid8huOqzhPw==,type:str]
    pgp:
        - created_at: "2021-03-25T16:36:22Z"
          enc: |
            -----BEGIN PGP MESSAGE-----

            hQIMAx/ZFgZfXuJNAQ//a6Soh/F6mEoxwW8jcVy4NxZUXW/aQJ16uNQur8xkuuHN
            AH15urUTBd8xN+3zQTuM0jun10G8cGOcdA0yBGeEQe8Axsh/ZNYkwijj2LjTpk8k
            7PGlSKALqZsOJ6hKl9VEOQ3+h1LBG6gFh4s916e/AYXXGd8pYi0kJiRGDclkt8U0
            4BotCNlaQ1BE67223SX9rjdVEAbs1lMgL/37+J3AEnIN12FdovOeXKxsHxtpEP2A
            a3uKeWGMW5FDPH9udTmlDiW5ekVRMAOggc7Ihtd74P5IwnSdbEZsZom0q49jv6/7
            VZ+IYKIpTpnfevgfEFmSDxOODGTp2X5btP+51jSvVl/CNKrWYCVvBudhHRuB3uHT
            vUp41n2hRznyq9dph9x3TEobaGwj3jokl+jwa6vc7DoeNHfPWsA2yyH5TavJinxP
            pufACD0XCfrL9+79yEWVt/IZDgIVupxjSIEz5+ID275Mx/5/XpHDwHlhTX/gXR0a
            Fssj6rm/1C4VJftZfFIkMOm1I/MoHUim1Syu/BL4j0WemZ+uI+RJVaxxzx0Q6Tyl
            EQHf3ZlK/Nzu4SqMiS4If7vwyYhRmZ6vMK9pHzObv1Zjbd5edueydD4tEzJoAA2r
            hHiPBIRkqLjYFdjmkm2DnzS0y+fFHdUoOi6hbRkPH5V4ULa/ThBEy/SzOWY89ADS
            XAElbOOdk5J5hImU+ZFS1zQLeOasf4Nwdbnd1jKG+aWBzlD8iUnsa2BpRqsGollj
            1kkZthi89CiQ7iknAWo5Yez0Y65p8IF0IPJ8XW46KP0jDbpjxQBHSFUu6g18
            =MW5w
            -----END PGP MESSAGE-----
          fp: 041173C69061E4F841DD8E080650CAE18738112D
    encrypted_regex: ^(data|stringData)$
    version: 3.7.0
```

I installed sops-gpg secret with pgp key into flux-system namespace:  

```plaintext:
gpg --export-secret-keys --armor "${KEY_FP}" |
kubectl create secret generic sops-gpg \
--namespace=flux-system \
--from-file=sops.asc=/dev/stdin
```

I already have flux git source setup with name 'podinfo' which points to flux-test repo, so I just need push busybox-configmap.yaml to git repo.

Next I'll setup busybox values.yaml with some default environment values which I will overwrite with Configmap values, I'll add these lines to values.yaml:  

```plaintext:
env:
  name1: first
  value1: "5"
  name2: second
  value2: "6"
  name3: third
  value3: "7"
```

busybox templates/deployment.yaml setup:  

```plaintext:
     env:
          - name: {{ .Values.env.name1 }} 
            value: {{ .Values.env.value1 | quote }} 
          - name: {{  .Values.env.name2 }} 
            value: {{ .Values.env.value2 | quote }} 
          - name: {{ .Values.env.name3 }}
            value: {{ .Values.env.value3 | quote }}
```

Complete busybox helmrelease config:  

```plaintext:
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: busybox
spec: {}
status: {}
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
kind: Kustomization
metadata:
  name: busyboxconfigmap
  namespace: flux-system 
spec:
  decryption:
    provider: sops
    secretRef:
      name: sops-gpg
  interval: 5m0s
  path: ./apps/kustomize/test1/busybox
  prune: true
  targetNamespace: flux-system
  sourceRef:
    kind: GitRepository
    name: podinfo
  validation: client
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: busybox
  namespace: flux-system
spec:
  interval: 10m0s
  chart:
    spec:
      chart: ./apps/base/charts/busybox/
      sourceRef:
        kind: GitRepository
        name: podinfo
      interval: 1m0s
  targetNamespace: busybox 
  valuesFrom:
    - kind: ConfigMap
      name: busyboxdata
      valuesKey: data.txt
```

Above config will:  

* create new namespace 'busybox'
* decrypt and deploy configmap 'busyboxdata' into flux-system namespace
* deploy my custom helm release 'busybox' into busybox namespace
* get helm deployment overwite values from configmap 'busyboxdata' which will replace values from values.yaml

My complete repo structure:  

```plaintext:
.
├── apps
│   ├── base
│   │   ├── charts
│   │   │   └── busybox
│   │   │       ├── charts
│   │   │       ├── Chart.yaml
│   │   │       ├── templates
│   │   │       │   ├── deployment.yaml
│   │   │       │   ├── _helpers.tpl
│   │   │       │   ├── hpa.yaml
│   │   │       │   ├── NOTES.txt
│   │   │       │   ├── service.yaml
│   │   │       │   └── tests
│   │   │       │       └── test-connection.yaml
│   │   │       └── values.yaml
│   │   └── podinfo
│   │       ├── deployment.yaml
│   │       ├── hpa.yaml
│   │       ├── kustomization.yaml
│   │       └── service.yaml
│   └── kustomize
│       ├── test1
│       │   ├── busybox
│       │   │   ├── busybox-configmap.yaml
│       │   │   └── data.txt
│       │   └── podinfo
│       │       ├── hpa.yaml
│       │       └── kustomization.yaml
│       └── test2
├── clusters
│   ├── test1
│   │   ├── apps1-kustomization.yaml
│   │   ├── busybox-helm.yaml
│   │   ├── flux-system
│   │   │   ├── gotk-components.yaml
│   │   │   ├── gotk-monitoring.yaml
│   │   │   ├── gotk-sync.yaml
│   │   │   └── kustomization.yaml
│   │   └── test1-source.yaml
│   └── test2
└── README.md
```

I have completed setup only for cluster 'test1' which has two apps: 1) podinfo with kustomize and 2) busybox with sops and custom helm charts. Similar way I could define more different type of clusters test2, test3 and so on.  

## GitOps infrastructure automation with Ansible and Flux

For cluster 'test2' I will create complete setup for my Go restapi platform:  

* deploy flux
* deloy krew
* deploy sops
* deloy helm
* deloy node labels
* deploy MetalLB loadbalancer
* deploy haproxy ingress
* deploy Prometheus
* deploy haproxy-fix
* deploy application layer: postgres and gorestapi Pods

## Kubernetes disaster recovery, how to re-install cluster state

If things get badly wrong, sometimes fastest way back is to re-install cluster state.  

kubeadm reset: clean up files that were created by kubeadm init or join. When executed in control-plane node, wipes out all info from previous cluster and print out join info to new cluster.  

kubeadm init: Initailize new cluster state. New config files will be created.  

kebeadm join: You have to re-join all worker nodes by executing kudeadm reset + kubeadm join.  

In my setup I did reset on all nodes, kubeadm init on control-plane and join on worker nodes. I also had to execute some additional commands in every node:  
cat <<EOF | sudo tee /etc/modules-load.d/crio.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

All labels are wiped out, so I had to re-label my cluster nodes for haproxy-ingress.  

I got this error: "failed to set bridge addr: "cni0" already has an IP address" for some starting Pod. I executed below commands in every node:  
sudo ip link set cni0 down  
sudo brctl delbr cni0  

Finally I rebooted control-plane node and cluster was back in business.  
