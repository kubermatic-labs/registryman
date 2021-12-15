#lang racket/base

(require racket/contract)
(require racket/match)
(require racket/set)
(require racket/dict)

(provide (contract-out
          [status-equal? (-> any/c any/c boolean?)]
          [remove-capabilities (-> hash? hash?)]))

(module+ test
  (require rackunit))

(define (remove-capabilities status-hsh)
  (foldl (λ (k v hsh)
           ;; (hash-set hsh k (hash-remove! v "capabilities"))
           (if (equal? k "capabilities")
               hsh
               (hash-set hsh k (if (hash? v)
                                   (remove-capabilities v)
                                   v))))
         (hash)
         (hash-keys status-hsh)
         (hash-values status-hsh)))

(define (status-equal? a b)
  (let ([equals (equal? (status-transform a) (status-transform b))])
    equals))
(module+ test
  (check-true (status-equal? 'a 'a))
  (check-true (status-equal? 1 1))
  (check-false (status-equal? 0 1))
  (check-false (status-equal? 1 "1"))
  (check-true (status-equal? #t #t))
  (check-true (status-equal? #f #f))
  (check-false (status-equal? #t #f))
  (check-true (status-equal? '() '()))
  (check-true (status-equal? '(1 2 3) '(1 2 3)))
  (check-true (status-equal? '(2 3 1) '(1 2 3)))
  (check-true (status-equal? '(3 2 1) '(1 2 3)))
  (check-true (status-equal? '((1 2) (3 4) (5 6)) '((1 2) (3 4) (5 6))))
  (check-true (status-equal? '((1 2) (3 4) (5 6)) '((1 2) (5 6) (3 4))))
  (check-true (status-equal? '((1 2) (3 4) (5 6)) '((2 1) (6 5) (3 4))))
  (check-false (status-equal? '() '(1)))
  (check-false (status-equal? '(1) '()))
  (check-true (status-equal? (hash) (hash)))
  (check-true (status-equal? (hash 'a 1) (hash 'a 1)))
  (check-false (status-equal? (hash 'a 1) (hash 'a 2)))
  (check-false (status-equal? (hash 'a 1) (hash 'b 1)))
  (check-false (status-equal? (hash 'a 1) (hash 'a "1")))
  (check-false (status-equal? (hash 'a 1) (hash 'a "1")))
  (check-true (status-equal? (hash 'a 1 'b 2) (hash 'a 1 'b 2)))
  (check-true (status-equal? (hash 'a 1 'b 2) (hash 'b 2 'a 1)))
  (check-true (status-equal? (hash 'a 1 'b '(1 2)) (hash 'b '(1 2) 'a 1)))
  (check-true (status-equal? (hash 'a 1 'b '(1 2)) (hash 'b '(2 1) 'a 1)))
  (check-true (status-equal? (hash 'a 1 'b '(1 2)) (hash 'b '(2 1) 'a 1)))
  (check-true (status-equal?
               #hash((members . ()) (name . global-images) (replicationrules . ()) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0))
               #hash((members . ()) (name . global-images) (replicationrules . ()) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0))
                             ))
  (check-true (status-equal?
               #hash((test-harbor . #hash((capabilities . #hash((cancreateproject . #t) (candeleteproject . #t) (canmanipulateprojectmembers . #t) (canmanipulateprojectreplicationrules . #t) (canmanipulateprojectscanners . #t) (canpullreplicate . #t) (canpushreplicate . #t) (hasprojectmembers . #t) (hasprojectreplicationrules . #t) (hasprojectscanners . #t) (hasprojectstoragereport . #t))) (projects . (#hash((members . ()) (name . global-images) (replicationrules . (#hash((direction . Push) (remoteregistryname . test-harbor2) (trigger . #hash((schedule . "") (type . event_based)))) #hash((direction . Push) (remoteregistryname . test-harbor2) (trigger . #hash((schedule . "") (type . event_based)))) #hash((direction . Push) (remoteregistryname . test-harbor2) (trigger . #hash((schedule . "") (type . event_based)))))) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0)))))) (test-harbor2 . #hash((capabilities . #hash((cancreateproject . #t) (candeleteproject . #t) (canmanipulateprojectmembers . #t) (canmanipulateprojectreplicationrules . #t) (canmanipulateprojectscanners . #t) (canpullreplicate . #t) (canpushreplicate . #t) (hasprojectmembers . #t) (hasprojectreplicationrules . #t) (hasprojectscanners . #t) (hasprojectstoragereport . #t))) (projects . (#hash((members . ()) (name . global-images) (replicationrules . ()) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0)) #hash((members . ()) (name . local-project) (replicationrules . ()) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0)))))))
               #hash((test-harbor . #hash((capabilities . #hash((cancreateproject . #t) (candeleteproject . #t) (canmanipulateprojectmembers . #t) (canmanipulateprojectreplicationrules . #t) (canmanipulateprojectscanners . #t) (canpullreplicate . #t) (canpushreplicate . #t) (hasprojectmembers . #t) (hasprojectreplicationrules . #t) (hasprojectscanners . #t) (hasprojectstoragereport . #t))) (projects . (#hash((members . ()) (name . global-images) (replicationrules . (#hash((direction . Push) (remoteregistryname . test-harbor2) (trigger . #hash((schedule . "") (type . event_based)))) #hash((direction . Push) (remoteregistryname . test-harbor2) (trigger . #hash((schedule . "") (type . event_based)))) #hash((direction . Push) (remoteregistryname . test-harbor2) (trigger . #hash((schedule . "") (type . event_based)))))) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0)))))) (test-harbor2 . #hash((capabilities . #hash((cancreateproject . #t) (candeleteproject . #t) (canmanipulateprojectmembers . #t) (canmanipulateprojectreplicationrules . #t) (canmanipulateprojectscanners . #t) (canpullreplicate . #t) (canpushreplicate . #t) (hasprojectmembers . #t) (hasprojectreplicationrules . #t) (hasprojectscanners . #t) (hasprojectstoragereport . #t))) (projects . (#hash((members . ()) (name . global-images) (replicationrules . ()) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0)) #hash((members . ()) (name . local-project) (replicationrules . ()) (scannerstatus . #hash((name . Trivy) (url . http://harbor-harbor-trivy:8080))) (storageused . 0)))))))
               ))
  
  (check-true (status-equal?
               #hash(("test-harbor" . #hash(("capabilities" . #hash(("cancreateproject" . #t) ("candeleteproject" . #t) ("canmanipulateprojectmembers" . #t) ("canmanipulateprojectreplicationrules" . #t) ("canmanipulateprojectscanners" . #t) ("canpullreplicate" . #t) ("canpushreplicate" . #t) ("hasprojectmembers" . #t) ("hasprojectreplicationrules" . #t) ("hasprojectscanners" . #t) ("hasprojectstoragereport" . #t))) ("projects" . (#hash(("members" . ()) ("name" . "global-images") ("replicationrules" . (#hash(("direction" . "Push") ("remoteregistryname" . "test-harbor2") ("trigger" . #hash(("schedule" . "") ("type" . "event_based")))) #hash(("direction" . "Push") ("remoteregistryname" . "test-harbor2") ("trigger" . #hash(("schedule" . "") ("type" . "event_based")))) #hash(("direction" . "Push") ("remoteregistryname" . "test-harbor2") ("trigger" . #hash(("schedule" . "") ("type" . "event_based")))))) ("scannerstatus" . #hash(("name" . "Trivy") ("url" . "http://harbor-harbor-trivy:8080"))) ("storageused" . 0)))))) ("test-harbor2" . #hash(("capabilities" . #hash(("cancreateproject" . #t) ("candeleteproject" . #t) ("canmanipulateprojectmembers" . #t) ("canmanipulateprojectreplicationrules" . #t) ("canmanipulateprojectscanners" . #t) ("canpullreplicate" . #t) ("canpushreplicate" . #t) ("hasprojectmembers" . #t) ("hasprojectreplicationrules" . #t) ("hasprojectscanners" . #t) ("hasprojectstoragereport" . #t))) ("projects" . (#hash(("members" . ()) ("name" . "global-images") ("replicationrules" . ()) ("scannerstatus" . #hash(("name" . "Trivy") ("url" . "http://harbor-harbor-trivy:8080"))) ("storageused" . 0)) #hash(("members" . ()) ("name" . "local-project") ("replicationrules" . ()) ("scannerstatus" . #hash(("name" . "Trivy") ("url" . "http://harbor-harbor-trivy:8080"))) ("storageused" . 0)))))))

               #hash(("test-harbor" . #hash(("capabilities" . #hash(("cancreateproject" . #t) ("candeleteproject" . #t) ("canmanipulateprojectmembers" . #t) ("canmanipulateprojectreplicationrules" . #t) ("canmanipulateprojectscanners" . #t) ("canpullreplicate" . #t) ("canpushreplicate" . #t) ("hasprojectmembers" . #t) ("hasprojectreplicationrules" . #t) ("hasprojectscanners" . #t) ("hasprojectstoragereport" . #t))) ("projects" . (#hash(("members" . ()) ("name" . "global-images") ("replicationrules" . (#hash(("direction" . "Push") ("remoteregistryname" . "test-harbor2") ("trigger" . #hash(("schedule" . "") ("type" . "event_based")))) #hash(("direction" . "Push") ("remoteregistryname" . "test-harbor2") ("trigger" . #hash(("schedule" . "") ("type" . "event_based")))) #hash(("direction" . "Push") ("remoteregistryname" . "test-harbor2") ("trigger" . #hash(("schedule" . "") ("type" . "event_based")))))) ("scannerstatus" . #hash(("name" . "Trivy") ("url" . "http://harbor-harbor-trivy:8080"))) ("storageused" . 0)))))) ("test-harbor2" . #hash(("capabilities" . #hash(("cancreateproject" . #t) ("candeleteproject" . #t) ("canmanipulateprojectmembers" . #t) ("canmanipulateprojectreplicationrules" . #t) ("canmanipulateprojectscanners" . #t) ("canpullreplicate" . #t) ("canpushreplicate" . #t) ("hasprojectmembers" . #t) ("hasprojectreplicationrules" . #t) ("hasprojectscanners" . #t) ("hasprojectstoragereport" . #t))) ("projects" . (#hash(("members" . ()) ("name" . "global-images") ("replicationrules" . ()) ("scannerstatus" . #hash(("name" . "Trivy") ("url" . "http://harbor-harbor-trivy:8080"))) ("storageused" . 0)) #hash(("members" . ()) ("name" . "local-project") ("replicationrules" . ()) ("scannerstatus" . #hash(("name" . "Trivy") ("url" . "http://harbor-harbor-trivy:8080"))) ("storageused" . 0)))))))
               ))
  )

(define/match (status-transform a)
  [((? list?)) (list->set (map status-transform a))]
  [((? hash?)) (foldl (λ (k v hsh)
                        (hash-set hsh k (status-transform v)))
                      (hash)
                      (hash-keys a)
                      (hash-values a))]
  [(_) a])

