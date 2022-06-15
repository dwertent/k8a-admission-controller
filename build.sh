#!/usr/bin/env bash
set -ex

# export ITAG=latest
export WTAG=v0

# dep ensure
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook .
chmod +x webhook
 
docker build --no-cache -f Dockerfile -t webhook:$WTAG .
#  docker push webhook:$WTAG
rm -rf webhook
 
# kubectl -n test patch  deployment webhook -p '{"spec": {"template": {"spec": { "containers": [{"name": "webhook", "imagePullPolicy": "Never"}]}}}}' || true
kubectl -n test set image deployment/webhook webhook=webhook:$WTAG || true 
kubectl delete pod -n test $(kubectl get pod -n test | grep webhook |  awk '{print $1}') || true 
kubectl logs -f -n test $(kubectl get pod -n test | grep webhook |  awk '{print $1}') || true      