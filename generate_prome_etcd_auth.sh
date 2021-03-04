#!/bin/bash

D="$(mktemp -d)"

sudo /bin/bash -s << EOF
cp /etc/kubernetes/pki/etcd/{ca.crt,healthcheck-client.{crt,key}} $D
chown -R joro:joro "$D"
EOF

kubectl create secret generic etcd-client --from-file="$D" -n monitoring

kubectl get secret etcd-client -n monitoring

rm -fr "$D"
