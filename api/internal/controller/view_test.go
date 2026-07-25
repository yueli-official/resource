package controller

import (
	"context"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	"platform/products/resource/api/internal/resourceauthz"
)

func TestIsAdminUsesInstanceProtectedSubjectNotIdentityRole(t *testing.T) {
	protected := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "site-operator"}
	module, err := authorization.NewMemory(
		authorization.MustCompile(resourceauthz.Definition()),
		authorization.MemoryOptions{
			RootScopeID:       resourceauthz.RootScopeID,
			ProtectedSubjects: []authorization.SubjectRef{protected},
			Constraints:       resourceauthz.ConstraintEvaluators(),
			Predicates:        resourceauthz.PredicateEvaluators(),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	service := resourceauthz.New(module, nil)

	operator := foundationauth.NewContext(context.Background(), &foundationauth.Principal{Subject: "site-operator"})
	operator = context.WithValue(operator, authorizationContextKey{}, service)
	if !isAdmin(operator) {
		t.Fatal("site operator rejected")
	}

	globalAdmin := foundationauth.NewContext(context.Background(), &foundationauth.Principal{Subject: "other", Roles: []string{"admin"}})
	globalAdmin = context.WithValue(globalAdmin, authorizationContextKey{}, service)
	if isAdmin(globalAdmin) {
		t.Fatal("global admin unexpectedly became resource operator")
	}
}
