#lang racket/base

(require racket/contract)
(require racket/generic)
(require testauto/parameters)
(require testauto/registry)
(require testauto/member)
(require testauto/resource)
(require testauto/scanner)
(require racket/hash)
(require racket/string)


(struct registry-project (name registries members scanner trigger)
  #:transparent
  #:methods gen:resource
  [
   (define/generic resource->yaml/generic resource->yaml)
   (define (resource->yaml resource)
      (hash "apiVersion" "registryman.kubermatic.com/v1alpha1"
            "kind" "Project"
            "metadata" (hash "name" (registry-project-name resource)
                             "namespace" (par:namespace))
            "spec" (hash-union
                    (hash "type" (registry-project-type resource)
                          "members" (map resource->yaml/generic
                                         (registry-project-members resource)))
                    (project-spec-localregistries resource)
                    (project-spec-scanner resource)
                    (project-spec-trigger resource))))

   (define (resource-filename resource)
     (format "~a-project.yaml" (registry-project-name resource)))
   (define (resource-deployment-priority resource)
     30)
   ])

(define (project-spec-trigger project)
  (let ([trig (registry-project-trigger project)])
    (if trig
        (let* ([trig-elems (string-split trig)]
               [trig-type (car trig-elems)])
          (cond [(string=? trig-type "cron")
                 (hash "trigger" (hash "type" trig-type
                                       "schedule" (string-join (cdr trig-elems))))]
                [else (hash "trigger" (hash "type" trig-type))]))
        (hash))))

(define (project-spec-scanner project)
  (let ([scn (registry-project-scanner project)])
    (if scn
       (hash "scanner" (scanner-name scn))
       (hash))))

(define (project-spec-localregistries project)
  (let ([registries (registry-project-registries project)])
    (if (null? registries)
        (hash)
        (hash "localRegistries" (map registry-name registries)))))

(define (registry-project-type project)
  (if (null? (registry-project-registries project))
      "Global"
      "Local"))

(define (project name
                 #:registries [registries '()]
                 #:members [members '()]
                 #:scanner [scanner #f]
                 #:trigger [trigger #f])
  (registry-project name registries members scanner trigger))

(provide (contract-out
          [project (->* (string?)
                        (#:registries (listof registry?)
                         #:members (listof project-member?)
                         #:scanner (or/c scanner? #f)
                         #:trigger (or/c string? #f))
                        registry-project?)]
          [struct registry-project ((name string?)
                                    (registries (listof registry?))
                                    (members (listof project-member?))
                                    (scanner (or/c scanner? #f))
                                    (trigger (or/c string? #f)))
            #:omit-constructor]))
