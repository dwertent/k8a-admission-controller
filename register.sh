#!/bin/bash
#version k8s-upload-register-template:144, Last update: 09:16:26 04/01/2022
export CA_NAMESPACE=test
export CA_NAMESPACE_TIER=test
export CA_LOGIN_SECRET_NAME=ca-login
export CA_IMAGE_REGISTRY=quay.io/armosec
export CA_WEBHOOK=webhook
export CA_CLUSTER_CERTIFICATE=cacert
export CA_WEBHOOK_PORT=443


if ! [ -x "$(command -v kubectl)" ]; then
  echo 'Error: kubectl not found, please install kubectl' >&2
  exit 1
fi

if ! [ -x "$(command -v wget)" ]; then
  echo 'Error: wget not found, please install wget' >&2
  exit 1
fi

if ! [ -x "$(command -v openssl)" ]; then
  echo 'Error: openssl not found, please install openssl' >&2
  exit 1
fi
 

set -e

DRYRUN="--dry-run=client"

cat <<EOF | kubectl  apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: ${CA_NAMESPACE}
  labels:
    app: ${CA_NAMESPACE}
    tier: ${CA_NAMESPACE_TIER}
---
kind: ServiceAccount
apiVersion: v1
metadata:
  labels:
    app: ca-controller-service-account
    tier: ${CA_NAMESPACE_TIER}
  name: ca-controller-service-account
  namespace: ${CA_NAMESPACE}

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ca-controller-roles
  labels:
    app: ca-controller-roles
    tier: ${CA_NAMESPACE_TIER}
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ca-controller-role-binding
  labels:
    app: ca-controller-role-binding
    tier: ${CA_NAMESPACE_TIER}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ca-controller-roles
subjects:
- kind: ServiceAccount
  name: ca-controller-service-account
  namespace: ${CA_NAMESPACE}

EOF

export csrName=${CA_WEBHOOK}.${CA_NAMESPACE}

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

export tmpdir=$(mktemp -d)

cat > ${tmpdir}/csr.conf <<EOF2 
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
prompt = no
[req_distinguished_name]
CN = ${CA_WEBHOOK}.${CA_NAMESPACE}.svc
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${CA_WEBHOOK}
DNS.2 = ${CA_WEBHOOK}.${CA_NAMESPACE}
DNS.3 = ${CA_WEBHOOK}.${CA_NAMESPACE}.svc
EOF2

CN_PATH="/"
UNAME=$(uname)
if [[ "$UNAME" == CYGWIN* || "$UNAME" == MINGW* ]] ; then
    CN_PATH=$CN_PATH$CN_PATH    
fi

openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -days 100000 -out ca.crt -subj "${CN_PATH}CN=admission_ca"

openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config ${tmpdir}/csr.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 100000 -extensions v3_req -extfile ${tmpdir}/csr.conf

kubectl  -n ${CA_NAMESPACE} delete secret ${CA_CLUSTER_CERTIFICATE} 2>/dev/null || true
kubectl  -n ${CA_NAMESPACE} create secret tls ${CA_CLUSTER_CERTIFICATE} --cert=server.crt --key=server.key ${DRYRUN} -o yaml | kubectl label -f- ${DRYRUN} -o yaml --local app=${CA_CLUSTER_CERTIFICATE} tier=${CA_NAMESPACE_TIER} | kubectl -n ${CA_NAMESPACE} apply -f -

export CA_BUNDLE=$(cat ca.crt | base64 | tr -d '\n')
export CRT_SUM=$(md5sum ca.crt)

rm ca.crt || true
rm ca.key || true
rm server.key || true
rm server.csr || true
rm server.crt || true
 

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: ${CA_WEBHOOK}
  namespace: ${CA_NAMESPACE}
  labels:
    app: ${CA_WEBHOOK}
spec:
  ports:
  - port: ${CA_WEBHOOK_PORT}
    targetPort: ${CA_WEBHOOK_PORT}
  selector:
    app: ${CA_WEBHOOK}
EOF

cat <<EOF3 | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${CA_WEBHOOK}
  namespace: ${CA_NAMESPACE}
  labels:
    app: ${CA_WEBHOOK}
    tier: ${CA_NAMESPACE_TIER}
spec:
  selector:         
    matchLabels:
      app: ${CA_WEBHOOK}
  replicas: 1
  template:
    metadata:
      labels:
        app: ${CA_WEBHOOK}
        tier: ${CA_NAMESPACE_TIER}
      annotations:
        certificate: "${CRT_SUM}"
    spec:
      containers:
        - name: ${CA_WEBHOOK}
          image: webhook:v0 
          imagePullPolicy: Never
          resources:
            requests:
              cpu: 300m
              memory: 100Mi
            limits:
              cpu: 1500m
              memory: 600Mi
          args:
            - -tlsCertFile=/etc/webhook/certs/tls.crt
            - -tlsKeyFile=/etc/webhook/certs/tls.key
            - -alsologtostderr
            - -v=4
            - 2>&1
          volumeMounts:
            - name: ${CA_CLUSTER_CERTIFICATE}
              mountPath: /etc/webhook/certs
              readOnly: true
      volumes:
        - name: ${CA_CLUSTER_CERTIFICATE}
          secret:
            secretName: ${CA_CLUSTER_CERTIFICATE}
      serviceAccountName: ca-controller-service-account      
EOF3

kubectl  delete ValidatingWebhookConfiguration ca-validate-cfg 2>/dev/null || true 
cat <<EOF4 | kubectl  create -f -
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: ca-validate-cfg
  namespace: ${CA_NAMESPACE}
  labels:
      app: ca-validate-cfg
      tier: ${CA_NAMESPACE_TIER}
webhooks:
- name: armo.validate.v1
  clientConfig:
    service:
      name: ${CA_WEBHOOK}
      namespace: ${CA_NAMESPACE}
      path: "/test"
    caBundle: ${CA_BUNDLE}
  timeoutSeconds: 30
  admissionReviewVersions: ["v1beta1", "v1"]
  sideEffects: None
  failurePolicy: Ignore  #Fail
  # objectSelector:
  #   matchExpressions:
  #   - key: armo.attach
  #     operator: In
  #     values: ["true", "false"]
  rules:
  - operations: [ "*" ]
    apiGroups: ["*"]
    apiVersions: ["*"]
    resources: ["*"]
  - operations: [ "*" ]
    apiGroups: ["*"]
    apiVersions: ["*"]
    resources: ["*"]
    scope: Namespaced
EOF4
