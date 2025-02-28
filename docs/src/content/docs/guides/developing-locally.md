---
title: Developing Locally
description: Learn how to develop Railpack locally
---

Once you've [checked out the repo](https://github.com/railwayapp/railpack), you
can follow this to start developing locally.

## Getting Setup

We use [Mise](https://mise.jdx.dev/) for managing language dependencies and
tasks for building and testing Railpack. You don't have to use Mise, but it's
recommended.

Install and use all versions of tools needed for Railpack

```bash
# Assuming you are cd'd into the repo root
mise install
```

Install all the Go dependencies

```bash
go mod tidy
```

List all the commands available

```bash
go run cmd/cli/main.go --help
```

## ① Building directly with Buildkit

**👋 Requirement**: an instance of Buildkit must be running locally.
Instructions in "[Run BuildKit Locally](#run-buildkit-locally)" at the bottom of
the readme.

Railpack will instantiate a BuildKit client and communicate to over GRPC in
order to build the generated LLB.

```bash
go run cmd/cli/main.go --verbose build examples/node-bun
```

You need to have a BuildKit instance running (see below).

## ② Custom frontend

You can build with a [custom BuildKit frontend](/guides/custom-frontend), but
this is a bit tedious for local iteration.

The frontend needs to be built into an image and accessible to the BuildKit
instance. To see how you can build and push an image, see the
`build-and-push-frontend` mise task in `mise.toml`.

Once you have an image, you can do:

Generate a build plan for an app:

```bash
go run cmd/cli/main.go plan examples/node-bun --out test/railpack-plan.json
```

Build the app with Docker:

```bash
docker build \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack:railpack-frontend" \
  -f test/railpack-plan.json \
  examples/node-bun
```

or use BuildKit directly:

```bash
buildctl build \
  --local context=examples/node-bun \
  --local dockerfile=test \
  --frontend=gateway.v0 \
  --opt source=ghcr.io/railwayapp/railpack:railpack-frontend \
  --output type=docker,name=test | docker load
```

_Note the `docker load` here to load the image into Docker. However, you can
change the [output](https://github.com/moby/buildkit?tab=readme-ov-file#output)
or push to a registry instead._

## Run BuildKit Locally

If building with the `build` command, you need to have a BuildKit instance
running with the `BUILDKIT_HOST` environment variable set to the container.

```bash
# Run a BuildKit instance as a container
docker run --rm --privileged -d --name buildkit moby/buildkit

# Set the buildkit host to the container
export BUILDKIT_HOST=docker-container://buildkit
```

## Mise commands

```bash
# Lint and format
mise run check

# Run tests
mise run test

# Start the docs dev server
mise run docs-dev
```
