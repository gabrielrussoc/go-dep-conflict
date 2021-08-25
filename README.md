# Go dependency conflicts

This is a simple repro to showcase how multiple go projects in a monorepo
can have conflicts in Bazel if they all have their own dependencies.

## Overview

The module `common` is a library that exports a function `common.F` that takes a `uuid` as an argument.

The module `service` is a binary that depends on `common` and calls `common.F`.

Both modules have their own `go.mod` files and declare their third party dependencies independently:

- `common` depends on `github.com/google/uuid v1.3.0`
- `service` depends on `github.com/google/uuid v1.2.0`

## The problem

If we `bazel build //service:service`, we get:

```
service/main.go:9:22: cannot use "databricks.com/service/vendor/github.com/google/uuid".New() (type "databricks.com/service/vendor/github.com/google/uuid".UUID) as type "databricks.com/common/vendor/github.com/google/uuid".UUID in argument to common.F
```

The type is obviously correct but they come from two different versions. In order to make the compiler happy, we need to make sure everyone is on the exact same version. Bazel has no concept of dependency resolution, so it is up to us to make sure the build is compatible.

**Update**: I realised they had different `importmap` attributes and that's why they failed. But if I use the same `importmap` I get:
```
link: package conflict error: vendor/github.com/google/uuid: multiple copies of package passed to linker:
	//service/vendor/github.com/google/uuid:uuid
	//common/vendor/github.com/google/uuid:uuid
```
## How to fix this

We create a single Go module named `third_party` that declares the dependencies of all of the projects. We then use Go to resolve those dependencies and 'bazelify' the result. Every go target will then depend on `//third_party/<lib>`. At Databricks, this is how we deal with Scala dependencies.

This has consequences. Adding new dependencies should be as simple as adding a new line to the shared `go.mod`. But as we upgrade dependencies, we have the potential to affect a lot of targets downstream, which can complicate things.

For Go in particular, there is a convention use a different import path for major versions: https://go.dev/blog/v2-go-modules
This could save us a lot of trouble since minor version upgrades tend to cause less trouble. And we can keep multiple major versions around if necessary.
