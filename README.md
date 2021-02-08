# gorestapi

My personal Go playground

## K8s installation:

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

## MetalLB loadbalancer installation:

kubectl get configmap kube-proxy -n kube-system -o yaml | \  
sed -e "s/strictARP: false/strictARP: true/" | \  
kubectl apply -f - -n kube-system  

kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.5/manifests/namespace.yaml  

kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.5/manifests/metallb.yaml  

kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"  

$ cat config.yaml  
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

kubectl apply -f config.yaml  

Change gorestapi-svc.yaml type: ClusterIP => type: LoadBalancer  

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