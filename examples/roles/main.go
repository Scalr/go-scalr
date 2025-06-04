package main

import (
	"context"
	"log"

	scalr "github.com/scalr/go-scalr"
)

func main() {
	config := scalr.DefaultConfig()
	client, err := scalr.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create a context
	ctx := context.Background()

	// Create a new role
	r, err := client.Roles.Create(ctx, scalr.RoleCreateOptions{
		Name:        scalr.String("example-role"),
		Description: scalr.String("This role is created from go-scalr "),
		Permissions: []*scalr.Permission{{ID: "*:*"}},
	})
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("The role with name %v was created. ID: %v", r.Name, r.ID)
	}

	// Update the role
	r, err = client.Roles.Update(ctx, r.ID, scalr.RoleUpdateOptions{
		Name:        scalr.String("go-scalr-role"),
		Permissions: []*scalr.Permission{{ID: "*:*"}, {ID: "global-scope:read"}},
	})
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("The role with id %v was updated. New permissions: %v", r.ID, []string{r.Permissions[0].ID, r.Permissions[1].ID})
	}

	// Delete the role
	err = client.Roles.Delete(ctx, r.ID)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("The role with id %v was deleted.", r.ID)
	}
}
