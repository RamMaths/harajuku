package http

import (
	"net/http"
	"strings"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"

	"github.com/gin-gonic/gin"
)

const (
	// authorizationHeaderKey is the key for authorization header in the request
	authorizationHeaderKey = "authorization"
	// authorizationType is the accepted authorization type
	authorizationType = "bearer"
	// authorizationPayloadKey is the key for authorization payload in the context
	authorizationPayloadKey = "authorization_payload"
)

// authMiddleware is a middleware to check if the user is authenticated
func authMiddleware(token port.TokenService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
      ctx.Next()
      return
    }

		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		isEmpty := len(authorizationHeader) == 0
		if isEmpty {
			err := domain.ErrEmptyAuthorizationHeader
			handleAbort(ctx, err)
			return
		}

		fields := strings.Fields(authorizationHeader)
		isValid := len(fields) == 2
		if !isValid {
			err := domain.ErrInvalidAuthorizationHeader
			handleAbort(ctx, err)
			return
		}

		currentAuthorizationType := strings.ToLower(fields[0])
		if currentAuthorizationType != authorizationType {
			err := domain.ErrInvalidAuthorizationType
			handleAbort(ctx, err)
			return
		}

		accessToken := fields[1]
		payload, err := token.VerifyToken(accessToken)

		if err != nil {
			handleAbort(ctx, err)
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

// adminMiddleware is a middleware to check if the user is an admin
func adminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
      ctx.Next()
      return
    }

		payload := getAuthPayload(ctx, authorizationPayloadKey)

		if payload == nil {
			handleAbort(ctx, domain.ErrForbidden)
			return
		}

		isAdmin := payload.Role == domain.Admin
		if !isAdmin {
			err := domain.ErrForbidden
			handleAbort(ctx, err)
			return
		}

		ctx.Next()
	}
}
