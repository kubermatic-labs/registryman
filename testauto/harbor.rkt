#lang rash

(require racket/contract)
(require racket/string)
(require yaml)
(require json)
(require testauto/registry)
(require testauto/resource)
(require testauto/kind-cluster)
(require net/url-string)
(require net/http-client)
(require net/base64)

(define harbor-release-name "harbor")
(define harbor-helm-repo "https://helm.goharbor.io")
(define harbor-values-file (let ([path (getenv "HARBOR_VALUES_FILE")])
                             (if (non-empty-string? path)
                                 path
                                 "harbor-values.yaml")))

(define harbor-helm-version-map (hash "2.3.3" "1.7.3"
                                      ;; "2.3.0" "1.7.0"
                                      "2.2.4" "1.6.4"
                                      ;; "2.2.3" "1.6.3"
                                      ;; "2.2.2" "1.6.2"
                                      ;; "2.2.1" "1.6.1"
                                      ;; "2.2.0" "1.6.0"
                                      ))

(define (harbor-supported-versions)
  (sort (hash-keys harbor-helm-version-map)
        string<?))

(define (harbor-helm-env-var-name harbor-version)
  (let ([chart-version (hash-ref harbor-helm-version-map harbor-version)])
    (format "HARBOR_HELM_~a" (string-replace chart-version "." "_" ))))

(struct harbor (cluster version namespace role)
  #:transparent
  #:methods gen:resource
  [
   (define resource->yaml registry-yaml)
   (define resource-filename registry-filename)
   (define (resource-deployment-priority registry)
     10)
   ]
  #:methods gen:registry
  [
   (define (registry-name registry)
     (format "~a-~a"
             (kind-cluster-name (harbor-cluster registry))
             (harbor-namespace registry)))

   (define (registry-role registry)
     (harbor-role registry))

   (define (registry-provider registry)
     "harbor")

   (define (registry-api-endpoint registry)
     (format "http://~a" (harbor-namespace registry)))

   (define (registry-username registry)
     "admin")

   (define (registry-password registry)
     "Harbor12345")

   (define (registry-install! registry)
     (displayln "Installing harbor")
     (with-kubeconfig (harbor-cluster registry)
       { helm install --kubeconfig (kubeconfig-path) --namespace (harbor-namespace registry) --create-namespace --values $harbor-values-file --set (format "externalURL=~a,expose.ingress.hosts.core=~a" (registry-api-endpoint registry) (harbor-namespace registry)) --wait harbor (getenv (harbor-helm-env-var-name (harbor-version registry)))})
     (harbor-ping registry))

   (define (registry-uninstall! registry)
     (with-kubeconfig (harbor-cluster registry)
       { helm uninstall $harbor-release-name --kubeconfig (kubeconfig-path) --namespace (harbor-namespace registry)})
     registry)])

(define (harbor-database-pod-name harbor)
  (with-kubeconfig (harbor-cluster harbor)
    #{ kubectl get pod --kubeconfig (kubeconfig-path) --namespace (harbor-namespace harbor) -l app=harbor,component=database (string->symbol "-o=jsonpath={.items[0].metadata.name}") }))

(define (harbor-clean-users! harbor)
  (with-kubeconfig (harbor-cluster harbor)
    { kubectl exec --kubeconfig (kubeconfig-path) (harbor-database-pod-name harbor) --namespace (harbor-namespace harbor) -c database -- psql -U postgres -d registry -c '"select * from harbor_user; delete from harbor_user where user_id > 2" }
    ))

(define (harbor-deployment-status harbor)
  (with-kubeconfig (harbor-cluster harbor)
    {helm status harbor -n (harbor-namespace harbor) --kubeconfig (kubeconfig-path) -o yaml |> read-yaml |>> hash-ref _ "info" |>> hash-ref _ "status" |>> string->symbol }) )

(define (harbor-list cluster [role #f])
  (map (lambda (chart)
         (harbor cluster (hash-ref chart "app_version") (hash-ref chart "namespace") role))
       (with-kubeconfig cluster
         { helm list --kubeconfig (kubeconfig-path) -A -f harbor -o yaml |> read-yaml })))

(define (harbor-ref cluster harbor-name role)
  (let ([h (findf (lambda (harbor)
                    (string=? (harbor-namespace harbor)
                              harbor-name))
                  (harbor-list cluster role))])
    (if h
        h
        (raise-user-error (format "harbor ~a does not exist in ~a"
                                  harbor-name
                                  (kind-cluster-name cluster))))))

(define (harbor-ping harbor)
  (let* ([host (url-host (string->url (registry-api-endpoint harbor)))]
         [b64auth (base64-encode (string->bytes/locale
                                  (format "~a:~a"
                                          (registry-username harbor)
                                          (registry-password harbor)))
                                 #"")]
         [headers (list (format "Authorization: Basic ~a" b64auth)
                        #"User-Agent: testauto"
                        #"Accept: */*")])
    (let-values ([(status-line _ body-port)
                  (http-sendrecv host
                                 "/api/v2.0/ping"
                                 #:method #"GET"
                                 #:headers headers
                                 #:port 80)])
      (case status-line
        [(#"HTTP/1.1 200 OK") (void)]
        [else (raise-user-error "error pinging harbor" status-line (read-string 1024 body-port))]))))

(define (harbor-provision-user! harbor username)
  (let* ([host (url-host (string->url (registry-api-endpoint harbor)))]
         [b64auth (base64-encode (string->bytes/locale
                                  (format "~a:~a"
                                          (registry-username harbor)
                                          (registry-password harbor)))
                                 #"")]
         [headers (list (format "Authorization: Basic ~a" b64auth)
                        #"Content-Type: application/json"
                        #"User-Agent: testauto"
                        #"Accept: */*")]
         [data (jsexpr->bytes (hash 'username username
                                    'comment (format "~a from testauto" username)
                                    'password "Pass1234"
                                    'realname "Test User"
                                    'email (format "~a@test.auto" username)))])
    (let-values ([(status-line _ body-port)
                  (http-sendrecv host
                                 "/api/v2.0/users"
                                 #:method #"POST"
                                 #:headers headers
                                 #:port 80
                                 #:data data)])
      (case status-line
        [(#"HTTP/1.1 201 Created") (void)]
        [else (raise-user-error "error creating user" status-line (read-string 1024 body-port))]))))

(provide (contract-out
          (harbor-supported-versions (-> (listof string?)))
          (harbor-list (->* (kind-cluster?) (registry-role?) (listof harbor?)))
          (harbor-ref (-> kind-cluster? string? registry-role? harbor?))
          (harbor-provision-user! (-> harbor? string? any/c))
          (harbor-clean-users! (-> harbor? any/c))
          (harbor-deployment-status (-> harbor? symbol?))
          [struct harbor ((cluster kind-cluster?)
                          (version string?)
                          (namespace string?)
                          (role registry-role?))]))
