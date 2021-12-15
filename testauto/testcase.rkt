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

(define (tc-registry-status-string tc #:in-cluster? [in-cluster? #f])
  (registry-status-string (tc-registries tc) #:in-cluster? in-cluster?)
  )

(define (tc-registry-expected-status-string tc)
  (in-resource-tmp-dir (testcase-resources tc)
                       #{ run-pipeline (par:registryman-path) status . -o yaml -e }))

(define (tc-apply! tc
                  #:dry-run? [dry-run? #f]
                  #:in-cluster? [in-cluster? #f])
  ;; (displayln "tc-apply!" dry-run?)
  (let ([dry-run-flag (if dry-run?
                          '--dry-run=true
                          '--dry-run=false)]
        [verbose-flag (if (par:verbose-mode)
                          '--verbose=true
                          '--verbose=false)])
    (if (or dry-run? (not in-cluster?))
        ;; (with-resources-deployed (testcase-resources tc) cluster
        ;;   ;; This could be used for CLI-with-deployed-resources
          ;; { run-pipeline (par:registryman-path) apply $dry-run-flag $verbose-flag --namespace (par:namespace) } 

        ;;   )
        (in-resource-tmp-dir (testcase-resources tc)
                             { run-pipeline (par:registryman-path) apply $dry-run-flag $verbose-flag . })
        (sleep 30) ;; Let's wait until operator stabilizes the state
        )))

(define (tc-print-resources tc)
  (for ([resource (testcase-resources tc)])
    (let ([filename (resource-filename resource)])
      (displayln "")
      (displayln filename)
      (displayln (make-string (string-length filename) #\-)))
    (write-yaml (resource->yaml resource)
                #:style 'block)))

(define (tc-upload-resources! tc cluster)
  (upload-resources! cluster (testcase-resources tc)))

(define (tc-delete-resources! tc cluster)
  (delete-resources! cluster (testcase-resources tc)))

(define (tc-validate tc [cluster #f])
  (in-resource-tmp-dir (testcase-resources tc)
                       { run-pipeline (par:registryman-path) validate . }))

(define (tc . resources)
  (testcase resources))

(define tc-resources testcase-resources)

(provide (contract-out
          (tc (->* () () #:rest (listof resource?) testcase?))
          (tc-resources (-> testcase? (listof resource?)))
          (tc-registries (-> testcase? (listof registry?)))
          (tc-registry-status-string (->* (testcase?) (#:in-cluster? boolean?) string?))
          (tc-registry-expected-status-string (-> testcase? string?))
          (tc-apply! (->* (testcase?)
                         (#:dry-run? boolean?
                          #:in-cluster? boolean?)
                         any/c))
          (tc-upload-resources! (-> testcase? kind-cluster? any/c))
          (tc-delete-resources! (-> testcase? kind-cluster? any/c))
          (tc-print-resources (-> testcase? any/c))
          (tc-validate (->* (testcase?) ((or/c kind-cluster? #f)) any/c))
          ))
