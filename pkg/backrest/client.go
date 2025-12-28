package backrest

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/garethgeorge/backrest/gen/go/v1"
	"github.com/garethgeorge/backrest/gen/go/v1/v1connect"
)

type Auth struct {
	BasicUsername string
	BasicPassword string
	BearerToken   string
}

type Client struct {
	backrest v1connect.BackrestClient
}

func New(baseURL string, auth Auth) *Client {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	interceptors := connect.WithInterceptors(authInterceptor(auth))
	return &Client{backrest: v1connect.NewBackrestClient(httpClient, baseURL, interceptors)}
}

func (c *Client) AddRepo(ctx context.Context, repo *v1.Repo) (*v1.Config, error) {
	resp, err := c.backrest.AddRepo(ctx, connect.NewRequest(repo))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

func authInterceptor(auth Auth) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if auth.BearerToken != "" {
				req.Header().Set("Authorization", "Bearer "+auth.BearerToken)
			} else if auth.BasicUsername != "" || auth.BasicPassword != "" {
				req.Header().Set("Authorization", "Basic "+basic(auth.BasicUsername, auth.BasicPassword))
			}
			return next(ctx, req)
		}
	}
}

func basic(user, pass string) string {
	// nolint: gosec // this is just encoding credentials.
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}
