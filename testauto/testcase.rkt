#lang rash

(require racket/contract)
(require racket/set)
(require racket/match)
(require racket/function)
(require yaml)
(require testauto/resource)
(require testauto/kind-cluster)
(require testauto/registry)
(require testauto/parameters)

(module+ test
  (require rackunit))

(struct testcase (resources)
  #:prefab)

(define (tc-registries tc)
  (filter registry? (testcase-resources tc)))

(define (tc-registry-status-string tc [cluster #f])
  (registry-status-string (tc-registries tc) #:cluster cluster))

(define (tc-registry-expected-status-string tc)
  (in-resource-tmp-dir (testcase-resources tc)
                       #{ run-pipeline (par:registryman-path) status . -o yaml -e }))

(define (tc-apply! tc
                  #:dry-run [dry-run #f]
                  #:cluster [cluster #f])
  (let ([dry-run-flag (if dry-run
                          '--dry-run=true
                          '--dry-run=false)]
        [verbose-flag (if (par:verbose-mode)
                          '--verbose=true
                          '--verbose=false)])
    (if cluster
        (with-resources-deployed (testcase-resources tc) cluster
          { run-pipeline (par:registryman-path) apply $dry-run-flag $verbose-flag --namespace (par:namespace) })
        (in-resource-tmp-dir (testcase-resources tc)
                             { run-pipeline (par:registryman-path) apply $dry-run-flag $verbose-flag . }))))

(define (tc-print-resources tc)
  (for ([resource (testcase-resources tc)])
    (let ([filename (resource-filename resource)])
      (displayln "")
      (displayln filename)
      (displayln (make-string (string-length filename) #\-)))
    (write-yaml (resource->yaml resource)
                #:style 'block)))

(define (tc-upload-resources! tc cluster)
  (upload-resources! (testcase-resources tc) cluster))

(define (tc-delete-resources! tc cluster)
  (delete-resources! (testcase-resources tc) cluster))

(define (tc-validate tc [cluster #f])
  (in-resource-tmp-dir (testcase-resources tc)
                       { run-pipeline (par:registryman-path) validate . }))

(define (tc . resources)
  (testcase resources))

(provide (contract-out
          (tc (->* () () #:rest (listof resource?) testcase?))
          (tc-registries (-> testcase? (listof registry?)))
          (tc-registry-status-string (->* (testcase?) ((or/c kind-cluster? #f)) string?))
          (tc-registry-expected-status-string (-> testcase? string?))
          (tc-apply! (->* (testcase?)
                         (#:dry-run boolean?
                          #:cluster (or/c kind-cluster? #f))
                         any/c))
          (tc-upload-resources! (-> testcase? kind-cluster? any/c))
          (tc-delete-resources! (-> testcase? kind-cluster? any/c))
          (tc-print-resources (-> testcase? any/c))
          (tc-validate (->* (testcase?) ((or/c kind-cluster? #f)) any/c))
          ))
