# Actions

Actions are the results of the `globalregistry.reconciler/Compare` function when
the expected and actual registry state are compared. Actions describe the
operations that are to be performed on the registry to reach the actual state.
Action is defined as an interface:

``` go
type Action interface {
	String() string
	Perform(globalregistry.Registry) (SideEffect, error)
}
```

An action has a string representation via the `String` method and and action can
be executed via the `Perform` method.

Examples for action implementation are:

  * creating a project
  * removing a member from a project
  * assign a scanner to a project, etc.

# SideEffects

On one hand Actions can succeed or fail, which is reflected in the returned
error. On the other hand, Actions may require further operations to be performed
which are not related to the registry. For example a file shall be stored in the
local file system. Such an operation is implemented as a SideEffect, which is
created when an Action is performed.

A SideEffect is an interface too:

``` go
type SideEffect interface {
	Perform(ctx context.Context) error
}
```

The parameters of the SideEffect shall be encapsulated into the `ctx` parameter
of the `Perform` method.

The `globalregistry/reconciler` package implements a `nilEffect` variable that
can be used as a no-op SideEffect.

<!-- Local Variables: -->
<!-- mode: gfm -->
<!-- End: -->
