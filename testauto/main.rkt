#lang racket/base

(require racket/cmdline)
(require racket/match)
(require racket/file)
(require racket/list)
(require testauto/kind-cluster)
(require testauto/testcase)
(require testauto/harbor)
(require testauto/registry)
(require testauto/status)
(require testauto/parameters)
(require yaml)

(define (parse-cluster-command cmd)
  (let ([subcommand-msg "Valid commands are:\n list\n create! <cluster-name> <cluster-version>\n delete! <cluster-name>\n supported-versions\n kubeconfig <cluster-name>\n deploy-crds! <cluster-name>\n import-registryman-image! <cluster-name>"])
    (match cmd
     [(list "list") (for ([cluster (kind-clusters)])
                      (displayln (format "- ~a [~a]"
                                         (kind-cluster-name cluster)
                                         (kind-cluster-version cluster))))]
     [(list "delete!" cluster-name) (kind-cluster-delete! cluster-name)]
     [(list "create!" cluster-name cluster-version) (kind-cluster-create!
                                                     (kind-cluster cluster-name cluster-version))]
     [(list "supported-versions") (for ([version (kind-supported-k8s-versions)])
                                    (displayln (format "- ~a" version)))]
     [(list "kubeconfig" cluster-name) (displayln (kind-cluster-kubeconfig (kind-clusters-ref cluster-name)))]
     [(list "import-registryman-image!" cluster-name)
      (kind-cluster-import-registryman-image! (kind-clusters-ref cluster-name))]
     [(list subcommand _ ...) (displayln (format "invalid sub-command for cluster: ~a~n~a"
                                                subcommand subcommand-msg
                                                ))]
     [(list) (displayln (format "cluster command requires a subcommand~n~a"
                                subcommand-msg))])))

(define (parse-harbor-command cmd)
  (let ([subcommand-msg "Valid commands are:\n versions\n list [<cluster-name>]\n install! <cluster-name> <harbor-name> <harbor-version>\n uninstall! <cluster-name> <harbor-name>\n add-user! <cluster-name> <harbor-name> <user-name>\n clean-users! <cluster-name> <harbor-name>"])
    (match cmd
      [(list "versions")
       (for ([version (harbor-supported-versions)])
         (displayln (format " ~a" version)))]
      [(list "list")
       (for ([cluster (kind-clusters)])
         (displayln (format "~a [~a]"
                            (kind-cluster-name cluster)
                            (kind-cluster-version cluster)))
         (for ([harbor (harbor-list cluster)])
           (displayln (format " ~a [~a]"
                              (harbor-namespace harbor)
                              (harbor-version harbor)))))]
      [(list "list" cluster-name)
       (for ([harbor (harbor-list (kind-clusters-ref cluster-name))])
         (displayln (format " ~a [~a]"
                            (harbor-namespace harbor)
                            (harbor-version harbor))))]
      [(list "install!" cluster-name harbor-name harbor-version)
       (let ([cluster (kind-clusters-ref cluster-name)])
         (registry-install! (harbor cluster
                                    harbor-version
                                    harbor-name
                                    'GlobalHub))
         (displayln "\n\nAdd the following line to your /etc/hosts file:\n")
         (displayln (format "~a ~a" (kind-cluster-ip cluster) harbor-name))
         (displayln (format "\nHarbor console is then at http://~a" harbor-name)))]

      [(list "uninstall!" cluster-name harbor-name)
       (registry-uninstall! (harbor (kind-clusters-ref cluster-name)
                                  ""
                                  harbor-name
                                  'GlobalHub))]

      [(list "add-user!" cluster-name harbor-name user-name)
       (harbor-provision-user! (harbor-ref (kind-clusters-ref cluster-name)
                                          harbor-name
                                          'Local)
                              user-name)
       (displayln (format "~a user created" user-name))]

      [(list "clean-users!" cluster-name harbor-name)
       (harbor-clean-users! (harbor-ref (kind-clusters-ref cluster-name) harbor-name 'Local))
       (displayln "user database is reset")]

      [(list subcommand _ ...) (displayln (format "invalid sub-command for harbor: ~a~n~a"
                                                  subcommand subcommand-msg))]
      [(list) (displayln (format "cluster command requires a subcommand~n~a"
                                 subcommand-msg))]))
  )

(define (tc-run tc-path cluster)
  (case (file-or-directory-type (simplify-path tc-path))
    [(file) (tc-run-file tc-path cluster)]
    [(directory) (tc-run-dir (simplify-path tc-path) cluster)]
    [(#f) (raise-user-error "path does not exist" tc-path)]
    [else (raise-user-error "path is neither a file nor a directory"
                            tc-path
                            (file-or-directory-type (simplify-path tc-path)))]))

(define (tc-run-dir tc-path cluster)
  (let ([paths (shuffle (find-files (λ (path)
                                      (not (equal? (file-or-directory-type path) 'directory)))
                                    tc-path
                                    #:skip-filtered-directory? #f))])
    (for ([path paths])
      (displayln (format "\n\n ######### ~a #########\n\n" path))
      (with-handlers ([exn:fail?
                       (λ (_)
                         (displayln "\n\n ######### EXECUTION FAILED ########\n\n")
                         (let ([executed-tests (append (takef paths
                                                              (λ (p) (not (equal? p path))))
                                                       (list path))])
                           (for ([start (cons "START" executed-tests)]
                                 [end executed-tests])
                             (let ([result (if (equal? end path)
                                               "FAILED"
                                               "SUCCEEDED")])
                               (displayln (format "~a -> ~a: ~a" start end result)))))
                         (raise-user-error "execution failed"))])
        (tc-run-file path cluster)))))

(define (tc-run-file tc-path cluster)
  (let* ([testcase (dynamic-require tc-path 'testcase)]
         [expected-status-string (tc-registry-expected-status-string testcase)]
         [expected-status (remove-capabilities (string->yaml expected-status-string))])
         (displayln (format "Testcase run of ~a" tc-path))
         (displayln "\n\n * * * RESOURCES * * *\n\n")
         (tc-print-resources testcase)
         (displayln "\n\n * * * VALIDATION * * *\n\n")
         (tc-validate testcase)
         (displayln "\n\n * * * EXPECTED STATUS * * *\n\n")
         (displayln expected-status-string)
         (displayln "\n\n * * * STATUS BEFORE EXECUTION * * *\n\n")
         (displayln (tc-registry-status-string testcase))
         (displayln "\n\n * * * DRY RUN * * *\n\n")
         (tc-apply! testcase
                    #:dry-run #t
                    #:cluster cluster)
         ;; (displayln "\n\n * * * CHECK ENVIRONMENT * * *\n\n")
         ;; ((dynamic-require tc-path 'check-status))
         ;; (displayln " SUCCESS! ")
         (displayln "\n\n * * * EXECUTE * * *\n\n")
         (tc-apply! testcase
                    #:dry-run #f
                    #:cluster cluster)
         (let* ([status-string (tc-registry-status-string testcase)]
                [status (remove-capabilities (string->yaml status-string))])
           (displayln "\n\n * * * STATUS AFTER EXECUTION * * *\n\n")
           (displayln status-string)
           (if (status-equal? status expected-status)
               (displayln "\n\n SUCCESS! \n\n")
               (begin
                 (displayln "\n\n FAILURE! \n\n")
                 (displayln "actual status does not match expected status")
                 (raise-user-error "execution failed"))))))

(define (parse-registryman-command cmd)
  (let ([subcommand-msg "Valid commands are:\n deploy! <cluster-name>\n delete! <cluster-name>\n log <cluster-name>"])
    (match cmd
      [(list "deploy!" cluster-name)
       (kind-cluster-deploy-registryman! (kind-clusters-ref cluster-name))]
      [(list "delete!" cluster-name)
       (kind-cluster-delete-registryman! (kind-clusters-ref cluster-name))]
      [(list "log" cluster-name)
       (kind-cluster-log-registryman! (kind-clusters-ref cluster-name))]
      [(list subcommand _ ...) (displayln (format "invalid sub-command for registryman: ~a~n~a"
                                                  subcommand subcommand-msg))]
      [(list) (displayln (format "registryman command requires a subcommand~n~a"
                                 subcommand-msg))])))

(define (parse-tc-command cmd)
  (let ([subcommand-msg "Valid commands are:\n print <tc-path>\n validate <tc-path>\n status <tc-path> [cluster-name]\n dry-run <tc-path> [cluster-name]\n apply <tc-path> [cluster-name]\n run <tc-path> [cluster-name]\n upload-resources! <tc-path>\n delete-resources! <tc-path>"])
    (match cmd
      [(list "print" tc-path)
       (tc-print-resources (dynamic-require (string->path tc-path ) 'testcase))]
      [(list "validate" tc-path)
       (tc-validate (dynamic-require (string->path tc-path) 'testcase))]
      [(list "status" tc-path)
       (display (tc-registry-status-string (dynamic-require (string->path tc-path) 'testcase)))]
      [(list "status" tc-path cluster-name)
       (display (tc-registry-status-string (dynamic-require (string->path tc-path) 'testcase)
                                           (kind-clusters-ref cluster-name)))]
      [(list "dry-run" tc-path)
       (tc-apply! (dynamic-require (string->path tc-path) 'testcase)
                  #:dry-run #t)]
      [(list "dry-run" tc-path cluster-name)
       (tc-apply! (dynamic-require (string->path tc-path) 'testcase)
                  #:dry-run #t
                  #:cluster (kind-clusters-ref cluster-name))]
      [(list "apply" tc-path)
       (tc-apply! (dynamic-require (string->path tc-path) 'testcase)
                  #:dry-run #f)]
      [(list "apply" tc-path cluster-name)
       (tc-apply! (dynamic-require (string->path tc-path) 'testcase)
                  #:dry-run #f
                  #:cluster (kind-clusters-ref cluster-name))]
      [(list "run" tc-path)
       (tc-run tc-path #f)]
      [(list "run" tc-path cluster-name)
       (tc-run tc-path (kind-clusters-ref cluster-name))]
      [(list "upload-resources!" tc-path cluster-name)
       (tc-upload-resources! (dynamic-require (string->path tc-path) 'testcase)
                             (kind-clusters-ref cluster-name))]
      [(list "delete-resources!" tc-path cluster-name)
       (tc-delete-resources! (dynamic-require (string->path tc-path) 'testcase)
                             (kind-clusters-ref cluster-name))]
      [(list subcommand _ ...) (displayln (format "invalid sub-command for tc: ~a~n~a"
                                                  subcommand subcommand-msg
                                                  ))]
      [(list) (displayln (format "tc command requires a subcommand~n~a"
                                 subcommand-msg))])))

; comment
(define (parse-command cmd cmd-rest)
  (match cmd
    ["cluster" (parse-cluster-command cmd-rest)]
    ["harbor" (parse-harbor-command cmd-rest)]
    ["registryman" (parse-registryman-command cmd-rest)]
    ["tc" (parse-tc-command cmd-rest)]
    [_ (displayln (format "invalid command: ~a~nValid commands are:~n cluster~n harbor~n tc~n registryman" cmd))]))

(module+ main
    (command-line #:program "testauto"
                  #:once-each
                  [("-v" "--verbose") "Compile with verbose messages"
                                      (par:verbose-mode #t)]
                  #:args (cmd . cmd-rest)
                  (parse-command cmd cmd-rest)))

