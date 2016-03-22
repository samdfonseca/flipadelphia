package server

import "net/http"

type MockAuth struct {
	OnAuthenticateRequest func(*http.Request) (bool, error)
}

func (mAuth MockAuth) AuthenticateRequest(r *http.Request) (bool, error) {
	return mAuth.OnAuthenticateRequest(r)
}
