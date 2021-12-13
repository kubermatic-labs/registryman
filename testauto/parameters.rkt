#lang racket/base

(define par:namespace (make-parameter "default"))
(define par:registryman-path (make-parameter (string->symbol
                                              (path->string (simplify-path
                                                             (let ([regman-env (or (getenv "REGISTRYMAN") "")])
                                                               (if (string=? regman-env "")
                                                                   "../../registryman"
                                                                   regman-env)))))))
(define par:verbose-mode (make-parameter #f))

(provide
 par:namespace
 par:registryman-path
 par:verbose-mode
 )
