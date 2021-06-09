# Registries and Projects

Registryman works on abstract resources called *registry* and *project*.

## Registry

The abstract registry resource corresponds to a real service that implements
Docker image registry HTTP API V2 [1]. On top of the Docker image registry API,
the service may implement further services for managing the repository
accesses, the replication rules, projects, etc.

A registry stores Docker images. The Docker images are grouped into
repositories, which help to identify a group of a collection of Docker images.

## Project

The abstract project resource is a container of repositories inside a registry.
A registry can have multiple projects, and each project can have multiple
repositories.

So the project layer groups repositories. Certain parameters of a repository can
be configured on project level which enables the configuration of multiple
repositories with a single operation in a convenient way.

For example an development organization has a single Docker registry but
multiple development teams. Each team has its own project. Each project contains
the repositories of the teams. The repository access can be configured on
project level by assigning members each project. 

# Global vs. Local

One of the main goal of the registryman project is to ease the configuration of
replication rules between the managed Docker registries and repositories.

To understand the concept of the replication provisioned by registry, we have to
introduce the following terms:

- global registry hub (only one can be configured)
- local registry
- global project
- local project

The global projects are projects that are automatically provisioned in all
registries: the global registry hub and all local registries.

The local projects are only provisioned in the configured registries.

Registryman provisions replication rules for global projects using the following
rule:

** Each repository of a global project is synchronized from the global registry
hub to the local registries. **

# Realizing Projects

Registryman provides support for multiple registry providers, i.e. multiple
implementations of the abstract registry resource. Some registry providers (e.g.
Harbor [2]) implement the concept of projects, some (e.g. ACR [3]) don't.

When the registry provider does not implement the concept of project,
registryman attempts to emulate the project concept based on the names of the
repositories.

E.g. in Harbor we have the following projects provisioned:

- `os-images`
- `databases`

Then we push the `alpine` and `ubuntu` images to the `os-images` project.
Similarly, we then push the `postgres` and `mongo` images to the `databases`
project.

In order to do so, we have to tag the images like this:

- `harbor.repo/os-images/alpine`
- `harbor.repo/os-images/ubuntu`
- `harbor.repo/databases/postgres`
- `harbor.repo/databases/mongo`

Sometimes we can't rely on the convenient project feature because the registry
does not support this concept. In that case we have to push the repositories
directly to the registry. In that case, we can (and should) came up with a good
naming strategy for our repositories. 
For ACR, we could push our 4 images tagged like this:

- `acr.repo/os-images/alpine`
- `acr.repo/os-images/ubuntu`
- `acr.repo/databases/postgres`
- `acr.repo/databases/mongo`

The repository names contain a "namespace" information (i.e.
`os-images` and `databases`).

This way we can emulate the project concept even if our registry provider does
not implement it.

Some risks and limitations of the project concept emulation are detailed in the
next sub-chapters.

## Creation of Projects

It can happen that registryman comes to the conclusion that it shall create a
new project because the expected state differs from the actual state of a
registry.

For registry providers with projects that's a meaningful administrative task. 

For registry providers without projects this is impossible to perform. Projects
are implicitly created when a repository is pushed with the proper tag. For
example to emulate the creation of the `runtimes` project in ACR, you should
push a repository called something like `acr.repo/runtimes/jvm`.

## Managing the Members of a Project

With projects we can manage the repository access on project level. We can e.g.
specify which users have push and which users have only pull capabilities.

Without projects, the repository access can be configured on registry level only.

Consequently, registryman ignores the project level membership configuration for
those registry providers that don't support the concept of projects.

## Replication

Even though the replication concept of registryman is based on projects, the
replication rule management may work even when a registry provider doesn't
implement the project feature.

E.g. there is an ACR registry called `acr-1` and a Harbor registry `harbor-1`. A
project called `databases` is configured so that it shall be synchronized from
`acr-1` to `harbor-1`.

In that case, registryman is able to provision a replication rule in `harbor-1`
with the following parameters:

- direction: `pull`
- remote registry: `acr-1`
- repository-filter: `databases/**`

[1]: https://docs.docker.com/registry/spec/api/
[2]: https://goharbor.io/docs/2.2.0/working-with-projects/
[3]: https://docs.microsoft.com/en-us/azure/container-registry/
