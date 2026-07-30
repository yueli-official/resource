package controller

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/yueli-official/foundation/go/authorization"

	"github.com/yueli-official/resource/api/internal/reserr"
	"github.com/yueli-official/resource/api/internal/resourceauthz"
)

type authorizationContextKey struct{}

func AuthorizationMiddleware(service *resourceauthz.Service) ghttp.HandlerFunc {
	return func(request *ghttp.Request) {
		ctx := context.WithValue(request.Context(), authorizationContextKey{}, service)
		correlationID := strings.TrimSpace(request.Header.Get("X-Trace-Id"))
		if correlationID == "" {
			correlationID = strings.TrimSpace(request.Header.Get("X-Request-Id"))
		}
		ctx = authorization.WithRequestMetadata(ctx, authorization.RequestMetadata{CorrelationID: correlationID})
		request.SetCtx(ctx)
		request.Middleware.Next()
	}
}

func authorizationService(ctx context.Context) *resourceauthz.Service {
	service, _ := ctx.Value(authorizationContextKey{}).(*resourceauthz.Service)
	return service
}

func isAdmin(ctx context.Context) bool {
	service := authorizationService(ctx)
	return service != nil && service.IsAdministrator(ctx)
}

func requireAdmin(ctx context.Context) error {
	return requireCapability(ctx, authorization.CapabilityManage, resourceauthz.RootScopeID, authorization.ResourceFacts{})
}

func requireCapability(
	ctx context.Context,
	capability authorization.CapabilityKey,
	scopeID authorization.ScopeID,
	resource authorization.ResourceFacts,
) error {
	service := authorizationService(ctx)
	if service == nil {
		return reserr.AuthorizationUnavailable()
	}
	decision, err := service.Decide(ctx, capability, scopeID, resource)
	if err != nil {
		if authorization.Is(err, authorization.ErrorUnavailable) {
			return reserr.AuthorizationUnavailable()
		}
		return reserr.Forbidden()
	}
	if !decision.Allowed {
		return reserr.Forbidden()
	}
	return nil
}

func authorizationResource(ctx context.Context, id string) (authorization.ResourceFacts, error) {
	service := authorizationService(ctx)
	if service == nil {
		return authorization.ResourceFacts{}, reserr.AuthorizationUnavailable()
	}
	resource, err := service.Resource(ctx, id)
	if err != nil {
		return authorization.ResourceFacts{}, mapAuthorizationError(err)
	}
	return resource, nil
}

func mapAuthorizationError(err error) error {
	switch {
	case authorization.Is(err, authorization.ErrorDenied):
		return reserr.Forbidden()
	case authorization.Is(err, authorization.ErrorUnavailable):
		return reserr.AuthorizationUnavailable()
	case authorization.Is(err, authorization.ErrorNotFound):
		return reserr.NotFound("authorization")
	case authorization.Is(err, authorization.ErrorInvalidInput),
		authorization.Is(err, authorization.ErrorConflict),
		authorization.Is(err, authorization.ErrorExpired):
		return reserr.InvalidInput(err.Error())
	default:
		return reserr.AuthorizationUnavailable()
	}
}
