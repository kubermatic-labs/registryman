#!/bin/bash

./ta-dev.sh cluster create! test 1.20
./ta-dev.sh cluster kubeconfig test > kubeconfig.yaml
./ta-dev.sh harbor install! test harbor 2.3.3
./ta-dev.sh harbor add-user! test harbor alpha
./ta-dev.sh harbor add-user! test harbor beta
./ta-dev.sh harbor install! test harbor2 2.3.3
./ta-dev.sh harbor add-user! test harbor2 alpha
./ta-dev.sh harbor add-user! test harbor2 beta
