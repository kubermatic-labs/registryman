#lang rash

(require racket/generic)
(require racket/contract)
(require yaml)
(require file/md5)
(require racket/file)

(define-generics resource
  [resource->yaml . (resource)]
  [resource-filename . (resource)]
  [resource-deployment-priority . (resource)])

(define (resource-deploy-ordering res1 res2)
  (< (resource-deployment-priority res1)
     (resource-deployment-priority res2)))

(define (resource-remove-ordering res1 res2)
  (> (resource-deployment-priority res1)
     (resource-deployment-priority res2)))

(define (md5-resources . resources)
  (let ([buf (open-output-string)])
    (for ([resource resources])
      (write-yaml (resource->yaml resource) buf))
    (substring (bytes->string/locale
                (md5 (open-input-string (get-output-string buf))))
               0 8)))

(define (testcase-directory-name testcase-name)
  (format "testcase-~a" testcase-name))

(define (resource-collect! dir . resources)
  (delete-directory/files dir
                          #:must-exist? #f)
  (make-directory* dir)
  (for ([resource resources])
    (write-yaml (resource->yaml resource)
                (open-output-file (format "~a/~a"
                                          dir
                                          (resource-filename resource)))
                #:style 'block))
  dir)

(define resource-dir (make-parameter ""))

(define-syntax-rule (in-resource-tmp-dir resources forms ...)
  (let ([tmp-dir-name (apply md5-resources resources)])
    (delete-directory/files tmp-dir-name
                            #:must-exist? #f)
    (make-directory* tmp-dir-name)
    (parameterize ([current-directory tmp-dir-name])
      (for ([resource resources])
        (with-output-to-file (resource-filename resource)
          (λ ()
            (write-yaml (resource->yaml resource)
                        #:style 'block)))))
    (let ([result (parameterize ([resource-dir tmp-dir-name]
                                 [current-directory tmp-dir-name])
                         forms ...)])
      (delete-directory/files tmp-dir-name
                              #:must-exist? #f)
      result)))


(provide gen:resource
         resource-dir
         in-resource-tmp-dir
         resource?
         (contract-out
          (resource->yaml (-> resource? yaml?))
          (resource-filename (-> resource? string?))
          (resource-deployment-priority (-> resource? real?))
          (resource-deploy-ordering (-> resource? resource? boolean?))
          (resource-remove-ordering (-> resource? resource? boolean?))
          (md5-resources (->* () () #:rest (listof resource?) string?))
          (resource-collect! (->* (string?) () #:rest (listof resource?) string?))))
