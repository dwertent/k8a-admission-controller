#!/bin/bash

export CA_NAMESPACE=test

kubectl delete ValidatingWebhookConfiguration ca-validate-cfg 2>/dev/null || true 
kubectl delete deployment webhook -n ${CA_NAMESPACE}
kubectl delete ns ${CA_NAMESPACE}
 