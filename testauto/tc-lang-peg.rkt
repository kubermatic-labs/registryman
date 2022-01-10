#lang racket/base

(require peg/peg)
(require racket/list)
(require racket/string)
(require syntax/strip-context)

(provide read read-syntax)

(define (read in)
  (syntax->datum (read-syntax #f in)))

(define (read-syntax src in)
  (set-box! expected-status "")
  (hash-clear! clusters)
  (hash-clear! harbors)
  (hash-clear! projects)
  (hash-clear! scanners)
  (peg source (read-string 65535 in))
  (generate-syntax))

(define-peg _ (* (or #\space #\tab)))
(define-peg _+ (+ (or #\space #\tab)))
(define-peg comment (and (char #\#)
                         (* (and (! (char #\newline)) (any-char)))))

;; (peg comment "# comment is comment")
;; (peg comment "# comment # is comment")

(define-peg id-char (or (range #\a #\z) (range #\A #\Z) (range #\0 #\9)(char #\_) (char #\-)))
(define-peg id (+ id-char))
(define-peg url-char (or id-char (char #\:) (char #\/) (char #\.)))
(define-peg url (+ url-char))

(define clusters (make-hash))
(define-peg cluster-definition (and (name cluster-id id)
                                    _
                                    (char #\=)
                                    _
                                    (string "k8s(")
                                    _
                                    (name cluster-name id)
                                    _
                                    (char #\)))
  (cond
    [(hash-has-key? clusters (string->symbol cluster-id))
     (raise-syntax-error #f (format "cluster ~a already defined" cluster-id))]
    [else (hash-set! clusters (string->symbol cluster-id) #`(kind-clusters-ref #,cluster-name))]))

;; (peg cluster-definition "cluster = k8s(cluster-test)")

(define harbors (make-hash))

(define-peg harbor-definition (and (name harbor-id id)
                                   _
                                   (char #\=)
                                   _
                                   (string "harbor(")
                                   _
                                   (name cluster-id id)
                                   _
                                   (char #\)))
  (define-harbor cluster-id harbor-id ''Local))

(define-peg global-harbor-definition (and (name harbor-id id)
                                   _
                                   (char #\=)
                                   _
                                   (string "global-harbor(")
                                   _
                                   (name cluster-id id)
                                   _
                                   (char #\)))
  (define-harbor cluster-id harbor-id ''GlobalHub))

(define (define-harbor cluster-id harbor-id role)
  (cond
    [(not (hash-has-key? clusters (string->symbol cluster-id)))
     (raise-syntax-error #f (format "k8s undefined: ~a" cluster-id))]
    [(hash-has-key? harbors (string->symbol harbor-id))
     (raise-syntax-error #f (format "harbor ~a already defined" harbor-id))]
    [else (hash-set! harbors (string->symbol harbor-id)
                     #`(harbor-ref #,(string->symbol cluster-id) #,harbor-id #,role))]))

;; (peg harbor-definition "test = harbor(cluster)")
;; (peg harbor-definition "test2 = harbor( cluster )")
;; (peg harbor-definition "test3=harbor( cluster)")
;; (peg global-harbor-definition "test4 = global-harbor(cluster)")

(define expected-status (box ""))

(define scanners (make-hash))

(define-peg scanner-definition (and (name scanner-id id)
                                   _
                                   (char #\=)
                                   _
                                   (string "scanner(")
                                   _
                                   (name scanner-url url)
                                   _
                                   (char #\)))
  (define-scanner scanner-id scanner-url))

(define (define-scanner scanner-id scanner-url)
  (cond
    [(hash-has-key? scanners (string->symbol scanner-id))
     (raise-syntax-error #f (format "scanner ~a already defined" scanner-id))]
    [else (hash-set! scanners (string->symbol scanner-id)
                     #`(scanner #,scanner-id #,scanner-url))]))

(define projects (make-hash))

(define-peg project-definition (and (name project-id id)
                                    _
                                    (char #\=)
                                    _
                                    (string "project(")
                                    _
                                    (char #\)))
  (cond
    [(hash-has-key? projects (string->symbol project-id))
     (raise-syntax-error #f (format "project ~a already defined" project-id))]
    [else (hash-set! projects (string->symbol project-id)
                     (hash))]))


;; (peg project-definition "local-project = project()")

(define-peg project-add-registry (and (name project-id id)
                                        (string ".registries")
                                        _
                                        (string "+=")
                                        _
                                        (name registry-id id))
  (cond
   [(not (hash-has-key? projects (string->symbol project-id)))
     (raise-syntax-error #f (format "unknown project: ~a" project-id))]
    [(not (hash-has-key? harbors (string->symbol registry-id)))
     (raise-syntax-error #f (format "unknown registry: ~a" registry-id))]
    [else (hash-set! projects (string->symbol project-id)
                     (let* ([proj (hash-ref projects (string->symbol project-id))]
                            [registries (hash-ref proj 'registries '())])
                       (hash-set proj 'registries (cons (string->symbol registry-id) registries))))]))

(define-peg project-add-member (and (name project-id id)
                                    (string ".members")
                                    _
                                    (string "+=")
                                    _
                                    (name member-name id)
                                    _+
                                    (string "as")
                                    _+
                                    (name member-role id))
  (cond
    [(not (hash-has-key? projects (string->symbol project-id)))
     (raise-syntax-error #f (format "unknown project: ~a" project-id))]
    [else (hash-set! projects (string->symbol project-id)
                    (let* ([proj (hash-ref projects (string->symbol project-id))]
                           [members (hash-ref proj 'members '())])
                      (hash-set proj 'members (cons (cons member-name member-role) members))))]))

(define-peg project-add-scanner (and (name project-id id)
                                     (string ".scanner")
                                    _
                                    (string "=")
                                    _
                                    (name scanner-id id))
  (cond
    [(not (hash-has-key? projects (string->symbol project-id)))
     (raise-syntax-error #f (format "unknown project: ~a" project-id))]
    [(not (hash-has-key? scanners (string->symbol scanner-id)))
     (raise-syntax-error #f (format "unknown scanner: ~a" scanner-id))]
    [else (hash-set! projects (string->symbol project-id)
                     (let* ([proj (hash-ref projects (string->symbol project-id))])
                       (hash-set proj 'scanner (string->symbol scanner-id))))]))

(define (project-define-syntax proj-id)
  (let ([project (hash-ref projects proj-id)])
    #`(define #,proj-id (project #,(symbol->string proj-id)
                                 #:registries (list #,@(hash-ref project 'registries '()))
                                 #:members (list #,@(for/list ([member (hash-ref project 'members '())])
                                                      #`(project-member #,(car member) #,(cdr member))))
                                 #:scanner #,(hash-ref project 'scanner #f)))))

;; (peg project-add-registry "local-project.registries += test")

(define-peg eof (! (any-char)))

(define-peg line (and _
                      (? (or cluster-definition
                             harbor-definition
                             global-harbor-definition
                             project-definition
                             project-add-registry
                             project-add-member
                             project-add-scanner
                             scanner-definition
                             ))
                      _
                      (? comment)
                      (char #\newline)))

;; (peg line "t = harbor(cluster)\n")
;; (peg line "t2 = harbor(cluster) # with a comment\n")
;; (peg line "c = k8s(test) # with a comment\n")
;; (peg line "# comment line\n")
;; (peg line "  # comment line\n")

(define-peg source (and (* line)
                        eof
                        ))

;; (peg source "a = k8s(test)\nh = harbor(a)")

(define (harbor-status-syntax)
  #`(define (check-harbor-status)
      (for ([harbor (list #,@(hash-keys harbors))])
        (unless (equal? 'deployed (harbor-deployment-status harbor))
          (raise-user-error "harbor status check failed"
                            (harbor-namespace harbor)
                            (kind-cluster-name (harbor-cluster harbor)))))))

(define (k8s-status-syntax)
  #`(define (check-k8s-status)
      (for ([cluster (list #,@(hash-keys clusters))])
        (kind-clusters-ref (kind-cluster-name cluster)))))

(define (generate-syntax)
  (strip-context #`(module tc racket/base
                     (require testauto/kind-cluster)
                     (require testauto/harbor)
                     (require testauto/project)
                     (require testauto/testcase)
                     (require testauto/scanner)
                     (require testauto/member)
                     (provide testcase check-status)
                     #,@(for/list ([cluster (hash->list clusters)])
                          #`(define #,(car cluster) #,(cdr cluster)))
                     #,@(for/list ([harbor (hash->list harbors)])
                          #`(define #,(car harbor) #,(cdr harbor)))
                     #,@(for/list ([scanner (hash->list scanners)])
                          #`(define #,(car scanner) #,(cdr scanner)))
                     #,@(for/list ([proj-id (hash-keys projects)])
                          (project-define-syntax proj-id))
                     #,(harbor-status-syntax)
                     #,(k8s-status-syntax)
                     (define (check-status)
                       (check-k8s-status)
                       (check-harbor-status))
                     (define testcase (tc #,@(append (hash-keys harbors)
                                                     (hash-keys scanners)
                                                     (hash-keys projects)))))))
