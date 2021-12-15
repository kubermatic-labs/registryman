#lang rash

(require racket/generic)
(require racket/contract)
(require yaml)
(require testauto/parameters)
(require testauto/resource)
(require testauto/kind-cluster)


(define-generics registry
  [registry-name . (registry)]
  [registry-role . (registry)]
  [registry-provider . (registry)]
  [registry-api-endpoint . (registry)]
  [registry-username . (registry)]
  [registry-password . (registry)]
  [registry-install! . (registry)]
  [registry-uninstall! . (registry)])

(define (registry-role? role-sym)
  (and (symbol? role-sym)
       (or (eq? role-sym 'GlobalHub)
           (eq? role-sym 'Local))))

(define (registry-yaml registry)
  (hash "apiVersion" "registryman.kubermatic.com/v1alpha1"
        "kind" "Registry"
        "metadata" (hash "name" (registry-name registry)
                         ;; "namespace" (registry-namespace registry))
                         "namespace" (par:namespace))
        "spec" (hash "role"  (symbol->string (registry-role registry))
                     "provider" (registry-provider registry)
                     "apiEndpoint" (registry-api-endpoint registry)
                     "username" (registry-username registry)
                     "password" (registry-password registry))))

(define (registry-filename registry)
  (format "~a-registry.yaml" (registry-name registry)))

(define (registry-status-string registries #:in-cluster? [in-cluster? #f])
  (let ([verbose-flag (if (par:verbose-mode)
                          '--verbose=true
                          '--verbose=false)])
    (if in-cluster?
        ;; (with-resources-deployed registries cluster
        ;;   #{ run-pipeline (par:registryman-path) status -o yaml $verbose-flag})
        #{ run-pipeline (par:registryman-path) status -o yaml $verbose-flag}
        (in-resource-tmp-dir registries
                             #{ run-pipeline (par:registryman-path) status . -o yaml $verbose-flag}))))

(provide gen:registry
         registry?
         registry-role?
         (contract-out
          (registry-filename (-> registry? string?))
          (registry-status-string (->* ((listof registry?)) (#:in-cluster? boolean?) string?))
          (registry-name (-> registry? string?))
          (registry-role (-> registry? registry-role?))
          (registry-provider (-> registry? string?))
          (registry-api-endpoint (-> registry? string?))
          (registry-username (-> registry? string?))
          (registry-password (-> registry? string?))
          (registry-install! (-> registry? any/c))
          (registry-uninstall! (-> registry? registry?))
          (registry-yaml (-> registry? yaml?))))
