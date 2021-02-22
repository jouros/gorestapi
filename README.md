# gorestapi

My personal Go playground.  

REST api written in Go and K8s deployment.  

## K8s installation

This K8s cluster is crio version from my other repo. Platform is Ubuntu 20.04 + KVM  

kubectl apply -f gorestapi-deployment.yaml  

kubectl apply -f gorestapi-svc.yaml  

cmd line api testing:  

curl -i http://127.0.0.1:8080/ping  

curl -i -H "Content-type: application/json" -d '{"title":"Hello","post":"World"}' http://127.0.0.1:8080/newsfeed  

curl -i http://127.0.0.1:8080/newsfeed  

K8s cmd line testing:  
kubectl get svc  
NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE  
gorestapi    ClusterIP   10.110.175.227   <none>        8080/TCP   4m55s  
kubernetes   ClusterIP   10.96.0.1        <none>        443/TCP    58d  

Check if app is responding:  
curl -i http://10.110.175.227:8080/ping  
HTTP/1.1 200 OK  
Content-Type: application/json; charset=utf-8  
Date: Mon, 08 Feb 2021 13:49:24 GMT  
Content-Length: 20  

{"hello":"Found me"}  

## MetalLB loadbalancer installation

kubectl get configmap kube-proxy -n kube-system -o yaml | \  
sed -e "s/strictARP: false/strictARP: true/" | \  
kubectl apply -f - -n kube-system  

kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.5/manifests/namespace.yaml  

kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.5/manifests/metallb.yaml  

kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"  

$ cat config.yaml  
```
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
NAME         TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE  
gorestapi    LoadBalancer   10.110.175.227   10.0.1.245    8080:32370/TCP   9m19s  
kubernetes   ClusterIP      10.96.0.1        <none>        443/TCP          58d  

Test from outside:  
  curl -i http://10.0.1.245:8080/ping
  HTTP/1.1 200 OK
  Content-Type: application/json; charset=utf-8
  Date: Mon, 08 Feb 2021 13:56:31 GMT
  Content-Length: 20

  {"hello":"Found me"}

## Install haproxy-ingress controller

More info: https://github.com/jcmoraisjr/haproxy-ingress/tree/master/examples/deployment

Label nodes 1-3 (I have 4 worker nodes):  
  kubectl label node worker1 role=ingress-controller
  node/worker1 labeled  
  kubectl label node worker2 role=ingress-controller
  node/worker2 labeled  
  kubectl label node worker3 role=ingress-controller
  node/worker3 labeled  

Check labels: 
kubectl get nodes --selector='role=ingress-controller'  
NAME      STATUS   ROLES    AGE   VERSION  
worker1   Ready    <none>   57d   v1.20.0  
worker2   Ready    <none>   57d   v1.20.0  
worker3   Ready    <none>   57d   v1.20.0  

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

curl -i http://10.0.1.248:8080/ping  
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
postgres     NodePort    10.100.109.205   <none>        5432:32531/TCP   28m  

Test db connection: https:  

psql -h 10.0.1.204 -U admin --password -p 32531 omadb  
Password:    
omadb=# quit  

## Wait for postgres connection with initContainer

Get script:  
git clone https://github.com/bells17/wait-for-it-for-busybox  

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

```apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
  - apiGroups:
      - "extensions"
      - "networking.k8s.io"
    resources:
      - ingresses
      - ingressclasses
      - ingresses/status```

Rsyslog sidecar for http access logs in haproxy-ingress-deployment.yaml:  

      ```- name: access-logs
        image: jumanjiman/rsyslog
        ports: 
        - name: udp
          containerPort: 514
          protocol: UDP```

