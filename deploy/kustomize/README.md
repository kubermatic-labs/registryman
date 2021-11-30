- [Create Key](#sec-1)
- [Create CSR](#sec-2)
- [Request the signing](#sec-3)
- [Sign the request](#sec-4)
- [Fetch the certificate](#sec-5)
- [Get the Kubernetes API server CA](#sec-6)
- [Deploy the resources](#sec-7)

# Create Key<a id="sec-1"></a>

```bash
openssl genrsa -out ca-key.pem 2048
```

# Create CSR<a id="sec-2"></a>

```bash
openssl req -new -key ca-key.pem -out csr.pem \
        -subj "/CN=registryman/O=system:nodes/CN=system:node:registryman.default.svc.cluster.local" \
        -addext "subjectAltName = DNS:registryman-webhook.default.svc" \
        -addext "extendedKeyUsage = clientAuth"
```

# Request the signing<a id="sec-3"></a>

```bash
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: registryman-webhook
spec:
  request: $(base64 -w 0 csr.pem)
  signerName: kubernetes.io/kubelet-serving
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF
```

# Sign the request<a id="sec-4"></a>

```bash
kubectl certificate approve registryman-webhook
```

# Fetch the certificate<a id="sec-5"></a>

```bash
kubectl get csr/registryman-webhook -o "jsonpath={.status.certificate}" | base64 -d - > ca.pem
```

# Get the Kubernetes API server CA<a id="sec-6"></a>

```bash
kubectl config view --raw --minify -o "jsonpath={.clusters[0].cluster.certificate-authority-data}"
```

# Deploy the resources<a id="sec-7"></a>

```bash
kustomize build . | kubectl apply -f -
```
