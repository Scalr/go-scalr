Scalr Go Client
==============================

## Installation

Installation can be done with a normal `go get`:

```
go get -u github.com/scalr/go-scalr
```

## Documentation

For complete usage of the API client, see the full [package docs](https://pkg.go.dev/github.com/scalr/go-scalr).

## Usage

```go
import scalr "github.com/scalr/go-scalr"
```

Construct a new Scalr client, then use the various endpoints on the client to
access different parts of the Scalr API. For example, to list
all environments:

```go
config := &scalr.Config{
	Token: "insert-your-token-here",
}

client, err := scalr.NewClient(config)
if err != nil {
	log.Fatal(err)
}

orgs, err := client.Environments.List(context.Background(), EnvironmentListOptions{})
if err != nil {
	log.Fatal(err)
}
```

## Examples

The [examples](https://github.com/Scalr/go-scalr/tree/master/examples) directory
contains a couple of examples. One of which is listed here as well:

```go
package main

import (
	"context"
	"log"

	scalr "github.com/scalr/go-scalr"
)

func main() {
	config := &scalr.Config{
		Token: "insert-your-token-here",
	}

	client, err := scalr.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create a context
	ctx := context.Background()

	// Create a new workspace
	w, err := client.Workspaces.Create(ctx, scalr.WorkspaceCreateOptions{
		Name: scalr.String("my-app-tst"),
		Environment: &scalr.Environment{ID: "env-ID"},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Update the workspace
	w, err = client.Workspaces.Update(ctx, w.ID, scalr.WorkspaceUpdateOptions{
		AutoApply:        scalr.Bool(false),
		TerraformVersion: scalr.String("0.12.0"),
		WorkingDirectory: scalr.String("my-app/infra"),
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

#Tests

You will need to set up the environment variables for your Scalr installation. For example:

```
export SCALR_ADDRESS=https://abcdef.scalr.com/
export SCALR_TOKEN=
```
You can run the acceptance tests like this:
```
make testacc
```