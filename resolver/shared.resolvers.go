package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"fmt"

	"github.com/albe2669/spotify-viewer/generated"
)

// Filler is the resolver for the filler field.
func (r *subscriptionResolver) Filler(ctx context.Context) (<-chan *bool, error) {
	panic(fmt.Errorf("not implemented: Filler - filler"))
}

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type subscriptionResolver struct{ *Resolver }