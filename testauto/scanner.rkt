#lang racket/base

(require racket/contract)
(require testauto/parameters)
(require testauto/resource)

(struct scanner (name url)
  #:transparent
  #:methods gen:resource
  [
   (define (resource->yaml resource)
     (hash "apiVersion" "registryman.kubermatic.com/v1alpha1"
           "kind" "Scanner"
           "metadata" (hash "name" (scanner-name resource)
                            "namespace" (par:namespace))
           "spec" (hash "url" (scanner-url resource))
           ))
   (define (resource-filename resource)
     (format "~a-scanner.yaml" (scanner-name resource)))
   (define (resource-deployment-priority scanner)
     20)
   ])

(provide (contract-out
          [struct scanner ((name string?)
                           (url string?))]))
