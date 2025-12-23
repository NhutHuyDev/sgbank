package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/NhutHuyDev/sgbank/internal/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationScheme = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authScheme := strings.ToLower(fields[0])
	if authScheme != authorizationScheme {
		return nil, fmt.Errorf("unsupported authorization type: %s", authScheme)
	}

	accessToken := fields[1]
	payload, err := server.TokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", authScheme)
	}

	return payload, nil
}
