#!/usr/bin/env bash

testauto cluster create! test 1.20
testauto cluster kubeconfig test >kubeconfig.yaml
testauto harbor install! test harbor 2.5.3
testauto harbor add-user! test harbor alpha
testauto harbor add-user! test harbor beta
testauto harbor install! test harbor2 2.3.3
testauto harbor add-user! test harbor2 alpha
testauto harbor add-user! test harbor2 beta
