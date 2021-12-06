#lang racket/base

(require racket/contract)
(require testauto/resource)

(struct project-member (name role)
  #:transparent
  #:methods gen:resource
  [
   (define (resource->yaml resource)
     (hash "name" (project-member-name resource)
           "role" (project-member-role resource)))
   ])

(provide (contract-out
          [struct project-member ((name string?)
                                  (role string?))]))
