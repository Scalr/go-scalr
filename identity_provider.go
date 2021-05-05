package scalr

// IdentityProvider represents a Scalr IACP IdentityProvider.
type IdentityProvider struct {
	ID string `jsonapi:"primary,identity-providers"`
}
