#lang rash

(require testauto/parameters)
(require racket/contract)
(require testauto/cert-manager)
(require testauto/resource)
(require racket/string)
(require file/md5)
(require json)
(require yaml)

(struct kind-cluster (name version)
  #:transparent)

(define (kind-version)
  #{ kind version |>> string-split |>> cadr })


(define kind-node-images
  #hash(("v0.11.1" . #hash(
                           ("1.21" . "kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6")
                           ("1.20" . "kindest/node:v1.20.7@sha256:cbeaf907fc78ac97ce7b625e4bf0de16e3ea725daf6b04f930bd14c67c671ff9")
                           ("1.19" . "kindest/node:v1.19.11@sha256:07db187ae84b4b7de440a73886f008cf903fcf5764ba8106a9fd5243d6f32729")
                           ("1.18" . "kindest/node:v1.18.19@sha256:7af1492e19b3192a79f606e43c35fb741e520d195f96399284515f077b3b622c")))))

(define (kind-supported-k8s-versions)
  (hash-keys (hash-ref kind-node-images (kind-version))))

(define (kind-node-image-of-version version)
  (let* ([k-version (kind-version)]
         [kind-image-hash (hash-ref kind-node-images k-version
                                   (λ ()
                                     (raise-user-error (format "Current kind version ~a is not supported" k-version))))])
    (hash-ref kind-image-hash version
              (λ ()
                (raise-user-error (format "Kubernetes version ~a is not supported. Valid versions are ~a"
                                          version
                                          (string-join (hash-keys kind-image-hash)
                                                       ", "
                                                       #:before-last " or ")))))))

(define (kind-node-version-of-image image)
  (let* ([k-version (kind-version)]
         [kind-image-hash (hash-ref kind-node-images k-version
                                    (λ ()
                                      (raise-user-error (format "Current kind version ~a is not supported" k-version))))])
    (ormap (λ (key-value)
             (let ([key (car key-value)]
                   [value (cdr key-value)])
               (if (string=? value image)
                   key
                   #f)))
            (hash->list kind-image-hash))))

(define (valid-kubernetes-version? version)
  (with-handlers ([exn:fail:user? (λ (_) #f)])
    (string? (kind-node-image-of-version version))))

(define cluster-apiVersion-string "kind.x-k8s.io/v1alpha4")
(define cluster-kind-string "Cluster")

(define kubeconfig-path (make-parameter ""))
(define kindconfig-path (make-parameter ""))

(define-syntax-rule (with-kindconfig cluster forms ...)
  (let* ([kindconfig-content (yaml->string (kind-cluster-yaml cluster)
                                           #:style 'block)]
         [kindconfig-filename (format "~a.kindconfig.yaml"
                                      (substring (bytes->string/locale
                                                  (md5 kindconfig-content))
                                                 0 8))])
    (with-output-to-file kindconfig-filename
      #:permissions #o600
      #:exists 'replace
      (λ () (display kindconfig-content)))
    (parameterize ([kindconfig-path kindconfig-filename])
      (let  ([result (with-handlers ([exn:fail?
                                      (λ (e)
                                        (delete-file kindconfig-filename)
                                        (raise e))])
                       (begin forms ...))])
        (delete-file kindconfig-filename)
        result))))


(define-syntax-rule (with-kubeconfig cluster forms ...)
  (let* ([kubeconfig-content (kind-cluster-kubeconfig cluster)]
         [kubeconfig-filename (format "~a.kubeconfig"
                                      (substring (bytes->string/locale
                                                  (md5 kubeconfig-content))
                                                 0 8))]
         [kubeconfig-exists? (file-exists? kubeconfig-filename)]
         [kubeconfig-env-value (or (getenv "KUBECONFIG") "")])
    (with-output-to-file kubeconfig-filename
      #:permissions #o600
      #:exists 'replace
      (λ () (display kubeconfig-content)))
    (when (getenv "DOCKER")
      { kubectl config set-cluster (string->symbol (format "kind-~a" (kind-cluster-name cluster))) (string->symbol (format "--server=https://~a-control-plane:6443" (kind-cluster-name cluster))) --insecure-skip-tls-verify=true --kubeconfig $kubeconfig-filename })
    (parameterize ([kubeconfig-path kubeconfig-filename])
      (let  ([result (with-handlers ([exn:fail?
                                      (λ (e)
                                        (delete-file kubeconfig-filename)
                                        (putenv "KUBECONFIG" kubeconfig-env-value)
                                        (raise e))])
                       (putenv "KUBECONFIG" kubeconfig-filename)
                       (begin forms ...))])

        (unless kubeconfig-exists?
          (delete-file kubeconfig-filename))
        (putenv "KUBECONFIG" kubeconfig-env-value)
        result))))

(define (kind-cluster-yaml cluster)
  (hash "kind" cluster-kind-string
        "apiVersion" cluster-apiVersion-string
        "name" (kind-cluster-name cluster)
        "nodes" (list (hash "role" "control-plane"
                            "kubeadmConfigPatches" (list (yaml->string (hash "kind" "InitConfiguration"
                                                                             "nodeRegistration" (hash "kubeletExtraArgs" (hash "node-labels" "ingress-ready=true")))))
                            "image" (kind-node-image-of-version (kind-cluster-version cluster))))))

(define (kind-cluster-create! cluster)
  (displayln "Creating kind cluster")
  (with-kindconfig cluster
   { kind create cluster --config (kindconfig-path) --kubeconfig .tmp.kubeconfig --wait 60s })
  { rm .tmp.kubeconfig }
  (with-kubeconfig cluster
    (displayln "Deploying nginx ingress controller")
    { kubectl apply -f (getenv "NGINX_DEPLOY") --kubeconfig (kubeconfig-path)}
    (sleep 20)
    (displayln "Waiting for nginx ingress controller")
    { kubectl wait --kubeconfig (kubeconfig-path) pod -l app.kubernetes.io/name=ingress-nginx,app.kubernetes.io/component=controller -n ingress-nginx --for condition=Ready --timeout=120s }
    (kind-cluster-deploy-cert-manager! cluster)
    (kind-cluster-deploy-registryman! cluster)
    ;; (kind-cluster-deploy-trivy! cluster)
    ))

(define (kind-cluster-delete! cluster-name)
  (let ([cluster (kind-clusters-ref cluster-name)])
    (with-kubeconfig cluster
      { kind delete cluster --name (kind-cluster-name cluster) --kubeconfig (kubeconfig-path)})))

(define (kind-clusters-ref cluster-name)
  (let* ([container-name #{ kind get nodes --name $cluster-name }]
         [version (with-handlers ([exn:fail?
                                   (λ (_)
                                     (raise-user-error (format "cluster ~a does not exist"
                                                               cluster-name)))])
                    { docker inspect $container-name |>> string->jsexpr |>> car |>> hash-ref _ 'Config |>> hash-ref _ 'Image |>> kind-node-version-of-image })])
    (kind-cluster cluster-name version)))

(define (kind-clusters)
  (let ([cluster-names {kind get clusters |>> string-split }])
    (map kind-clusters-ref cluster-names)))

(define (kind-cluster-ip cluster)
  { docker inspect (format "~a-control-plane" (kind-cluster-name cluster)) |> read-json |> car |> hash-ref _ 'NetworkSettings |> hash-ref _ 'Networks |> hash-ref _ 'kind |> hash-ref _ 'IPAddress })

(define (kind-cluster-kubeconfig cluster)
  #{ kind get kubeconfig --name (kind-cluster-name cluster)})

(define (kind-cluster-deploy-crds! cluster)
  (with-kubeconfig cluster
    { kubectl apply -f (getenv "PROJECT_CRD") --kubeconfig (kubeconfig-path)}
    { kubectl apply -f (getenv "REGISTRY_CRD") --kubeconfig (kubeconfig-path)}
    { kubectl apply -f (getenv "SCANNER_CRD") --kubeconfig (kubeconfig-path)}))

(define (kind-cluster-delete-crds! cluster)
  (with-kubeconfig cluster
    { kubectl delete -f (getenv "PROJECT_CRD") --kubeconfig (kubeconfig-path)}
    { kubectl delete -f (getenv "REGISTRY_CRD") --kubeconfig (kubeconfig-path)}
    { kubectl delete -f (getenv "SCANNER_CRD") --kubeconfig (kubeconfig-path)}))

(define (kind-cluster-import-registryman-image! cluster)
  (kind-cluster-import-image-archive-targz! cluster
                                            (getenv "REGISTRYMAN_DOCKER_IMAGE")))

(define (kind-cluster-import-image-archive-targz! cluster archive-path)
  { unpigz -c $archive-path | kind load image-archive /dev/stdin --name (kind-cluster-name cluster) })

(define (kind-cluster-deploy-cert-manager! cluster)
  (with-kubeconfig cluster
    { kubectl apply -f (getenv "CERT_MANAGER_YAML") --kubeconfig (kubeconfig-path)}
    { kubectl wait --kubeconfig (kubeconfig-path) pod -l app.kubernetes.io/component=webhook,app.kubernetes.io/instance=cert-manager -n cert-manager --for condition=Ready --timeout=120s })
  (upload-resources! cluster (list (cert-manager-clusterissuer "selfsigned"))))

(define (kind-cluster-deploy-trivy! cluster)
  (with-kubeconfig cluster
    { helm install trivy (getenv "TRIVY_HELM") -f (getenv "TRIVY_VALUES") --kubeconfig (kubeconfig-path) --wait}))

(define (kind-cluster-delete-trivy! cluster)
  (with-kubeconfig cluster
    { helm uninstall trivy --kubeconfig (kubeconfig-path)}))

(define (kind-cluster-deploy-registryman! cluster)
  (kind-cluster-import-registryman-image! cluster)
  (kind-cluster-deploy-crds! cluster)
  (let ([registryman-manifests (if (par:verbose-mode)
                                   (getenv "REGISTRYMAN_DEPLOYMENT_MANIFESTS_VERBOSE")
                                   (getenv "REGISTRYMAN_DEPLOYMENT_MANIFESTS"))])
   (with-kubeconfig cluster
     { kustomize build $registryman-manifests | kubectl apply -f - })))

(define (kind-cluster-delete-registryman! cluster)
  (with-kubeconfig cluster
    { kustomize build (getenv "REGISTRYMAN_DEPLOYMENT_MANIFESTS") | kubectl delete -f - })
  (kind-cluster-delete-crds! cluster))

(define (kind-cluster-log-registryman! cluster)
  (with-kubeconfig cluster
    { kubectl logs -n registryman deployment/registryman-webhook }))

(define (kubectl-on-resources! action resource-ordering)
  (lambda (cluster resources)
    (in-resource-tmp-dir resources
                         (with-kubeconfig cluster
                           (for ([filename (map resource-filename (sort resources
                                                                        resource-ordering))])
                             { kubectl $action -f $filename --kubeconfig (kubeconfig-path) --namespace (par:namespace)})))))

(define upload-resources! (kubectl-on-resources! 'apply resource-deploy-ordering))
(define delete-resources! (kubectl-on-resources! 'delete resource-remove-ordering))

(define-syntax-rule (with-resources-deployed resources cluster forms ...)
  (begin
    (upload-resources! cluster resources)
    (let ([result (with-handlers ([exn:fail? (λ (e)
                                               (delete-resources! cluster resources)
                                               (raise e))])
                    (with-kubeconfig cluster
                      (begin forms ...)))])
      (delete-resources! cluster resources)
      result)))

(provide
 kubeconfig-path
 with-resources-deployed
 with-kubeconfig
 (contract-out
  [kind-supported-k8s-versions (-> (listof string?))]
  [kind-cluster-create! (-> kind-cluster? any)]
  [kind-cluster-delete! (-> string? any)]
  [kind-clusters-ref (-> string? kind-cluster?)]
  [kind-clusters (-> (listof kind-cluster?))]
  [kind-cluster-ip (-> kind-cluster? string?)]
  [kind-cluster-kubeconfig (-> kind-cluster? string?)]
  ;; [kind-cluster-deploy-crds! (-> kind-cluster? any)]
  [kind-cluster-import-image-archive-targz! (-> kind-cluster? string? any)]
  [kind-cluster-import-registryman-image! (-> kind-cluster? any)]
  [kind-cluster-deploy-registryman! (-> kind-cluster? any)]
  [kind-cluster-delete-registryman! (-> kind-cluster? any)]
  [kind-cluster-log-registryman! (-> kind-cluster? any)]
  (upload-resources! (-> kind-cluster? (listof resource?) any/c))
  (delete-resources! (-> kind-cluster? (listof resource?) any/c))
  [struct kind-cluster ((name string?)
                        (version valid-kubernetes-version?))]))
