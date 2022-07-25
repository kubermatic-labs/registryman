#lang racket/base

(require peg)
(require racket/list)
(require racket/match)

(provide read read-syntax)

(define (read in)
  (syntax->datum (read-syntax #f in)))

(define (read-syntax src in)
  (peg l-source (read-string 65535 in)))

(define-peg eof (! (any-char)))

(define-peg/drop _ (* (or #\space #\newline)))

(define-peg/drop l-comment (and _ (char #\#)
                                (* (and (! (char #\newline)) (any-char)))
                                (or (char #\newline)
                                    eof)))

(define-peg l-symbol-char (or (range #\a #\z) (range #\A #\Z) (range #\0 #\9) #\-))
(define-peg l-symbol (name res (+ l-symbol-char)) (string->symbol res))

(define-peg l-number (name res (+ (range #\0 #\9))) (string->number res))
(define-peg l-string (and (char #\") (name res (* (and (! #\") (any-char)))) (char #\")) res)

(define-peg l-map-kv (and _ (name var l-symbol) _ (char #\=) _ (name val l-value) _ (char #\;)) (cons var val))

(define-peg l-map (and (char #\{) _ (name kvs (* l-map-kv)) _ (char #\})) kvs)

(define-peg l-array (and (char #\[) _ (name res (and (* (and l-value _ (drop (char #\, )) _)) (? l-value ))) _ #\]) `(list ,@res))

(define-peg l-kind (and (string "kind") #\space _ (name cluster-name l-value))
  `(kind-clusters-ref ,cluster-name))

(define-peg l-harbor (and (string "harbor")
                          #\space _
                          (name harbor-name l-value)
                          #\space _
                          (name cluster l-value)
                          (name is-global? (? (and (drop #\space) _ (string "global"))))
                          (name is-insecure? (? (and (drop #\space) _ (string "insecure"))))
                          )
  `(harbor-ref ,cluster
               ,harbor-name
               ,(case is-global?
                  [("global") ''GlobalHub]
                  [else ''Local])
               ,(case is-insecure?
                  [("insecure") '#t]
                  [else '#f])
               ))

(define-peg l-member (and (string "member") #\space _ (name member-name l-value) #\space _ (name member-role l-value))
  `(project-member ,member-name ,member-role)
  )

(define-peg l-scanner (and (string "scanner") #\space _ (name scanner-name l-value) #\space _ (name scanner-url l-value))
  `(scanner ,scanner-name ,scanner-url))

(define-peg l-project (and (string "project") #\space _ (name project-name l-value) (? (and #\space _ (name parameters l-map)) ))
  `(project ,project-name ,@(if parameters
                                            (foldl (match-lambda*
                                                     [(list (cons 'registries registries) result) (append (list '#:registries registries) result)]
                                                     [(list (cons 'members members) result) (append (list '#:members members) result)]
                                                     [(list (cons 'scanner scanner) result) (append (list '#:scanner scanner) result)]
                                                     [(list (cons unknown-key _) result) (raise-syntax-error 'syntax-error (format "unknown project parameter: ~a" unknown-key) )]
                                                     )
                                                   '()
                                                   parameters)
                                            '())
            ))

(define-peg/bake l-value (or l-kind l-harbor l-project l-scanner l-member
                             l-number l-string l-symbol l-map l-array))

(define-peg l-assignment (and (name var l-symbol) _ (char #\=) _ (name val l-value)) `(define ,var  ,val))

(define-peg l-expr (and _ (name res l-assignment ) _ (char #\;)) res)

(define (expression-defined-resource expr)
  (match expr
    [(list 'define var (list 'project _ ...)) var]
    [(list 'define var (list 'harbor-ref _ ...)) var]
    [(list 'define var (list 'scanner _ ...)) var]
    [_ #f]))

(define test-src #<<EOS
  # This is a test comment
  num = 15;
  name="project-name";
  c=kind "test";
  h=harbor "harbor" c;
  h2=harbor "harbor2" c global;
  h3=harbor "harbor3" kind "test";
  h4=harbor "harbor4" kind "test" insecure;
  h5=harbor "harbor5" c global insecure;
  a = "project-name";
  p=project a; # A comment at the end of the line
  pp=project "bla" {};
  p3=project "blabla" {
          registries = [h, h2];
  };
  p4=project "memberproject" {
          members = [];
  };
  p5= project "memberproject2" {
          members = [ member "alpha" "Developer" ];
  };

  p6= project "memberproject3" {
          members = [
            member "alpha" "Developer",
            member "beta" "Maintainer",
          ];
  };
  admin = member "gamma" "Administrator";
  p7 = project "memberproject4" {
          members = [
            member "alpha" "Developer",
            admin
          ];
  };
  p8 = project "memberproject4" {
          registries = [h];
          members = [
            admin
          ];
  };
  scanner1 = scanner "trivy" "http://trivi.com:8000";
  p9 = project "scannerproject" {
          registries = [h];
          members = [
            member "alpha" "Administrator"
          ];
          scanner = scanner1;
  };
  p10 = project "embeddedscanner" {
          registries = [h];
          members = [
            member "alpha" "Administrator"
          ];
          scanner = scanner "embedded" "http://embedded:3000";
  };
  # Comment
EOS
)

(define-peg l-source (and (name expressions (* (or l-comment l-expr))) _ eof)
  #`(module source racket/base
      (require testauto/kind-cluster)
      (require testauto/project)
      (require testauto/testcase)
      (require testauto/scanner)
      (require testauto/harbor)
      (require testauto/member)
      (provide testcase)
      #,@expressions
      (define testcase (tc #,@(filter-map expression-defined-resource expressions)))))
