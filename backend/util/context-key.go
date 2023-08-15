package util

import (
	"context"
	"fmt"

	"firebase.google.com/go/auth"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	ContextKeyUserToken  = contextKey("userToken")
	ContextKeyUserRecord = contextKey("userRecord")
)

// GetUserTokenFromContext gets the user token from context.
func GetUserTokenFromContext(ctx context.Context) (*auth.Token, bool) {
	userToken, ok := ctx.Value(ContextKeyUserToken).(*auth.Token)
	return userToken, ok
}

// GetUserRecordFromContext gets the user record from context.
func GetUserRecordFromContext(ctx context.Context) (*auth.UserRecord, error) {
	userToken, ok := ctx.Value(ContextKeyUserRecord).(*auth.UserRecord)
	if !ok {
		return nil, fmt.Errorf("could not get user record from context")
	}
	return userToken, nil
}
