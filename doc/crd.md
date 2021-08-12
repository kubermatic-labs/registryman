# Configuration stored as Custom Kubernetes Resources

Besides the configuration files stored in the local filesystem, Registryman is
able to read the configuration from Kubernetes. In this case the registries,
projects and scanners are stored as Custom Resources in a Kubernetes namespace.

# Deploy the Custom Resource Definitions

Before storing the Registry, Project and Scanner resources, we shall deploy the
Custom Resource Definitions (CRD).

```bash
kubectl apply -f pkg/apis/registryman/v1alpha1/registryman.kubermatic.com_registries.yaml \
              -f pkg/apis/registryman/v1alpha1/registryman.kubermatic.com_projects.yaml   \ 
              -f pkg/apis/registryman/v1alpha1/registryman.kubermatic.com_scanners.yaml
```

# Deploy Custom Resources

After the custom resources are deployed, we can deploy the Registry, Project and
Scanner resources.

```bash
kubectl apply -f examples/global-registry.yaml
kubectl apply -f examples/global-project.yaml
kubectl apply -f examples/scanner.yaml
```
