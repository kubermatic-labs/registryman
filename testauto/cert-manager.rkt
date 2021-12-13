#lang racket/base

(require testauto/resource)
(require racket/contract)

(define (cert-manager-clusterissuer->yaml ci)
  (hash "apiVersion" "cert-manager.io/v1"
        "kind" "ClusterIssuer"
        "metadata" (hash "name" (cert-manager-clusterissuer-name ci))
        "spec" (hash "selfSigned" (hash))))

(define (cert-manager-clusterissuer-filename ci)
  (format "~a-clusterissuer.yaml" (cert-manager-clusterissuer-name ci)))

(struct cert-manager-clusterissuer (name)
  #:transparent
  #:methods gen:resource
  [
   (define resource->yaml cert-manager-clusterissuer->yaml)
   (define resource-filename cert-manager-clusterissuer-filename)
   (define (resource-deployment-priority ci)
     5)
   ])

(provide (contract-out
          (struct cert-manager-clusterissuer ((name string?)))))
