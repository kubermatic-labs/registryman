
# Goal

Test automation tool for registryman aims to help with the automated testing of the registryman tool in a real Kubernetes environment.

Everything that can be tested using unit tests shall be tested with unit tests and not with the test automation tool.

# Prerequisites

It shall be possible to run test automation tool on developers' laptops, and it shall be possible to integrate it with any CI system. Due to the many external dependencies (Racket runtime, Go compiler, external code generator tools, etc.), the test automation tool strives to emit reproducible output. Thus, the tool and its environment are self-contained.

This is achieved with the help of [Nix](https://nixos.org/). Before using the test automation tool, you have to [install](https://nixos.org/download.html) Nix first.

Then, the test automation tool can be used from the `registryman/testauto` directory.

The test automation tool has been tested only on Linux operating system.

# Test Environment

The test automation tool takes care of the Kubernetes test environment. It uses `kind` in the background.

## Listing the supported Kubernetes versions

Test automation tool supports different versions of Kubernetes. To list the supported Kubernetes versions, enter:

```bash
./ta-dev.sh cluster supported-versions
```

## Starting a new test environment

You can start a new test environment with the following command:

```bash
./ta-dev.sh cluster create! env-name 1.20
```

A new environment will be created with the name `env-name` and version `1.20`.

## Listing the test environments

You can list the previously created test environments with the following command:

```bash
./ta-dev.sh cluster list
```

## Get an environment's kubeconfig file

The `kubeconfig` of an environment can be generated with the following command:

```bash
./ta-dev.sh cluster kubeconfig env-name
```

You can store the kubeconfig file of the environment `test-env` with the following command:

```bash
./ta-dev.sh cluster kubeconfig test-env > kubeconfig
```

## Destroy a test environment

You can tear down a test environment with the following command:

```bash
./ta-dev.sh cluster delete! env-name
```

# Managing Harbor deployments

The test automation tool can manage the lifecycle of Harbor deployments of a test environment.

## Listing supported Harbor versions

The test automation tool can show the supported Harbor versions with this command:

```bash
./ta-dev.sh harbor versions
```

## Installing Harbor

You can deploy Harbor in a previously created test environment with the following command:

```bash
./ta-dev.sh harbor install! env-name harbor-name 2.2.4
```

Here, the `env-name` is the name of the previously created test environment, `harbor-name` is the name of the Harbor deployment and `2.2.4` is the Harbor version to deploy.

The `harbor-name` shall be unique across all managed test environments.

After Harbor is deployed, the test automation tool tells us how to update the /etc/hosts file.

```text
Add the following line to your /etc/hosts file:

172.18.0.3 harbor-name

Harbor console is then at http://harbor-name
```

## Listing the Harbor deployments

You can list the Harbor deployments of a test environments using the following command:

```bash
./ta-dev.sh harbor list env-name
```

Here, the `env-name` is the name of the test environment.

If you omit the env-name, Harbor deployments of all test environments are listed:

```bash
./ta-dev.sh harbor list
```

## Adding user to Harbor deployment

The test automation tool lets you create users in the Harbor user database with the following command:

```bash
./ta-dev.sh harbor add-user! env-name harbor-name user-name
```

## Cleaning up the Harbor user database

You can clean up the Harbor user database with the following command:

```bash
./ta-dev.sh harbor clean-users! env-name harbor-name
```

## Uninstalling Harbor

You can uninstall a previously installed Harbor with the following command:

```bash
./ta-dev.sh harbor uninstall! env-name harbor-name
```

Here, the `env-name` is the name of the previously created test environment, `harbor-name` is the name of the previously installed Harbor deployment.

# Registryman Deployment

The test automation tool helps you with the deployment of the following `registryman` components:

-   CRDs
-   validation webhook
-   operator

When a new test environment is created, the `registryman` components are automatically deployed. The `registryman` components are built and generated from the assets of the `registryman` folder, i.e. the test automation tool's parent directory.

The commands below detect changes in the source code of registryman. If you change the `registryman` source code, and then execute deploy command, the previously deployed `registryman` will be upgraded to a new version. The deployment involves the building of the binary, containerization and Kubernetes deployment. As such, the test automation tool can be considered as a basic tool for the development pipeline.

## Deploying CRDs

The `registryman` CRDs can be (re-)deployed using the following command:

```bash
./ta-dev.sh registryman deploy-crds! env-name
```

## Deploying registryman

The `registryman` operator and validating webhook components can be (re-)deployed with the following command:

```bash
./ta-dev.sh registryman deploy! env-name
```

## Checking the Logs of registryman Validation Webhook Deployment

You can check the logs of the registryman validation webhook container with the following command:

```bash
./ta-dev.sh registryman log env-name
```

## Deleting registryman deployment

The `registryman` operator and validating webhook components can be removed from the test environment with the following command:

```bash
./ta-dev.sh registryman delete! env-name
```

# Running tests

The main goal of the test automation tool is to run tests. The tests are defined in the tools own simple test language. For the current tests check the `registryman/testauto/tests` directory.

Since `registryman` can be used both as a CLI tool and a Kubernetes operator, the testing follows the same duality.

## CLI Mode overview

## Operator Mode overview

## Printing the YAML files of a testcase

You can print the yaml files of testcase with the following command:

```bash
./ta-dev.sh tc print tests/tc2.tc
```

The last argument of the command denotes the testcase under investigation.

## Validating the YAML files of a testcase

You can validate (with `registryman validate`) the yaml files of testcase with the following command:

```bash
./ta-dev.sh tc validate tests/tc2.tc
```

The last argument of the command denotes the testcase under investigation.

## Dry-run application of YAML files (CLI mode)

You can execute the `registryman apply --dry-run` command for a testcase with the following command:

```bash
./ta-dev.sh tc dry-run tests/tc2.tc
```

You can turn on verbose logging with:

```bash
./ta-dev.sh -v tc dry-run tests/tc2.tc
```

## Application of YAML files (CLI mode)

You can execute the `registryman apply` command for a testcase with the following command:

```bash
./ta-dev.sh tc apply tests/tc2.tc
```

You can turn on verbose logging with:

```bash
./ta-dev.sh -v tc apply tests/tc2.tc
```

## Uploading the Resources of a Testcase

You can upload the YAML files of a testcase to the test environment as Custom Resources with the following command:

```bash
./ta-dev.sh tc upload-resources! tests/tc2.tc env-name
```

## Deleting the Resources of a Testcase

You can delete the Custom Resources of a testcase from the test environment with the following command:

```bash
./ta-dev.sh tc delete-resources! tests/tc2.tc env-name
```

## Executing a Single Testcase (CLI mode)

You can execute a single testcase using the CLI mode using the following command:

```bash
./ta-dev.sh tc run tests/tc2.tc
```

The following steps are performed when a test is run in CLI mode:

1.  The resources are printed as YAML files.
2.  The resources are validated. See `registryman validate`.
3.  The expected status is printed. See `registryman status -e`.
4.  The actual status is printed. See `registryman status`.
5.  A dry-run is performed. See `registryman apply --dry-run`.
6.  The test is executed. See `registryman apply`.
7.  The actual status is printed again. See `registryman status`.

If the actual status of step 7 is the same as the expected status of step 3, the test is considered as successful. Otherwise, it is considered as failed.

You can turn on verbose logging:

```bash
./ta-dev.sh -v tc run tests/tc2.tc
```

## Executing a Batch of Testcases (CLI mode)

When you specify a directory name as the path to the testcase, all testcases within the specified directory will be executed:

```bash
./ta-dev.sh tc run tests
```

You can turn on verbose logging:

```bash
./ta-dev.sh -v tc run tests
```

## Executing a Single Testcase (Operator mode)

You can execute a single testcase using the `registryman` operator using the following command:

```bash
./ta-dev.sh tc run tests/tc2.tc env-name
```

Here, the `env-name` denotes the test environment, where the `registryman` operator is run.

The following steps are performed when a test is run in operator mode:

1.  The resources are printed as YAML files.
2.  The resources are validated. See `registryman validate`.
3.  The expected status is printed. See `registryman status -e`.
4.  The actual status is printed. See `registryman status`.
5.  The YAML resources are deployed.
6.  Waiting for a given time. The operator performs the changes
7.  The actual status is printed again. See `registryman status`.
8.  Deleting the YAML resources.

If the actual status of step 7 is the same as the expected status of step 3, the test is considered as successful. Otherwise, it is considered as failed.

You can turn on verbose logging:

```bash
./ta-dev.sh -v tc run tests/tc2.tc env-name
```

## Executing a Batch of Testcases (Operator mode)

When you specify a directory name as the path to the testcase, all testcases within the specified directory will be executed:

```bash
./ta-dev.sh tc run tests env-name
```

Here, the `env-name` denotes the test environment, where the `registryman` operator is run.

You can turn on verbose logging:

```bash
./ta-dev.sh -v tc run tests env-name
```
