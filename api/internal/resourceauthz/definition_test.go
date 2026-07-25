package resourceauthz_test

import (
	"context"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	"platform/products/resource/api/internal/resourceauthz"
)

func TestDefinitionAutomaticallyGrantsContributorAndEnforcesOwnership(t *testing.T) {
	admin := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "admin"}
	module, err := authorization.NewMemory(
		authorization.MustCompile(resourceauthz.Definition()),
		authorization.MemoryOptions{
			RootScopeID: resourceauthz.RootScopeID, ProtectedSubjects: []authorization.SubjectRef{admin},
			Constraints: resourceauthz.ConstraintEvaluators(), Predicates: resourceauthz.PredicateEvaluators(),
		},
	)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	ctx := context.Background()
	user := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "contributor"}
	result, err := module.ReconcileSubject(ctx, authorization.ReconcileSubjectCommand{Subject: user})
	if err != nil || result.Created != 1 || result.Grants[0].Role != resourceauthz.RoleContributor {
		t.Fatalf("ReconcileSubject() = %#v, %v", result, err)
	}
	if _, err := module.RegisterScope(ctx, authorization.RegisterScopeCommand{
		ID: resourceauthz.ResourceScopeID("item-1"), Type: resourceauthz.ScopeResource,
		ParentID: resourceauthz.RootScopeID,
	}); err != nil {
		t.Fatalf("RegisterScope() error = %v", err)
	}
	resource := resourceauthz.OwnedResource("item-1", user.ID)
	decision, err := module.Decide(ctx, authorization.DecisionRequest{
		Subject: user, Capability: resourceauthz.CapabilityItemPublish,
		ScopeID: resourceauthz.ResourceScopeID("item-1"), Resource: resource,
	})
	if err != nil || !decision.Allowed {
		t.Fatalf("Decide(own publish) = %#v, %v; want allow", decision, err)
	}
	resource.Relations[resourceauthz.RelationOwner] = []authorization.SubjectRef{{
		Kind: authorization.SubjectUser, ID: "other",
	}}
	decision, err = module.Decide(ctx, authorization.DecisionRequest{
		Subject: user, Capability: resourceauthz.CapabilityItemPublish,
		ScopeID: resourceauthz.ResourceScopeID("item-1"), Resource: resource,
	})
	if err != nil || decision.Allowed {
		t.Fatalf("Decide(other publish) = %#v, %v; want deny", decision, err)
	}
	service := resourceauthz.New(module, nil)
	userContext := foundationauth.NewContext(ctx, &foundationauth.Principal{Subject: user.ID})
	access, err := service.EffectiveAccess(userContext)
	if err != nil || len(access.Grants) == 0 {
		t.Fatalf("EffectiveAccess() = %#v, %v", access, err)
	}
}
