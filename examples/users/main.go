package main

import (
	"context"
	"log"
	"os"

	scalr "github.com/scalr/go-scalr"
)

func main() {
	config := scalr.DefaultConfig()
	config.Headers.Set("Prefer", "profile=internal")

	client, err := scalr.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create a context
	ctx := context.Background()

	// Create a user
	opts := scalr.UserCreateOptions{
		Email:            scalr.String("i.kovalkovskyi6@scalr.com"),
		IdentityProvider: &scalr.IdentityProvider{ID: os.Getenv("IDP_ID")},
		Status:           scalr.UserStatusActive,
	}

	usr, err := client.Users.Create(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("User created %v", usr.ID)

	// Retrieve a user
	usr, err = client.Users.Read(ctx, usr.ID)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("User %v was created at: %v", usr.ID, usr.CreatedAt)

	// Retrieve all users
	users, err := client.Users.List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("There are %v users on the server", len(users.Items))

	// Update the user
	usr, err = client.Users.Update(ctx, usr.ID, scalr.UserUpdateOptions{
		FullName: scalr.String("Ivan Kovalkovskyi"),
		Status:   scalr.UserStatusInactive,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Update status of the user %v: %v", usr.ID, usr.Status)

	// Delete the user
	err = client.Users.Delete(ctx, usr.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Deleted user %v", usr.ID)

}
