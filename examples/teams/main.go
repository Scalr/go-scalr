package main

import (
	"context"
	"log"

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

	users := []*scalr.User{{ID: "user-suh84u6vuvidtbg"}}

	// Create a team
	opts := scalr.TeamCreateOptions{
		IdentityProvider: &scalr.IdentityProvider{ID: "idp-sohkb0o1phrdmr8"},
		Account:          &scalr.Account{ID: "acc-svrcncgh453bi8g"},
		Name:             scalr.String("GoTeams4"),
		Description:      scalr.String("Team created by go-scalr"),
		Users:            users, // test@scalr.com
	}

	team, err := client.Teams.Create(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Team created %v", team.ID)

	// Retrieve a team
	team, err = client.Teams.Read(ctx, team.ID)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Team %v Description: %v", team.ID, team.Description)

	// Retrieve all teams
	teams, err := client.Teams.List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("There are %v teams on the server", len(teams.Items))

	// Update the team
	team, err = client.Teams.Update(ctx, team.ID, scalr.TeamUpdateOptions{
		Description: scalr.String("Edited"),
		//		Name:        scalr.String("Edited"),
		//		Users:       []*scalr.User{},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Update Description of the team %v: %v", team.ID, team.Description)

	// Delete the team
	err = client.Teams.Delete(ctx, team.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Deleted team %v", team.ID)

}
