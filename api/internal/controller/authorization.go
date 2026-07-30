package controller

import (
	"context"

	"github.com/yueli-official/foundation/go/authorization"

	v1 "github.com/yueli-official/resource/api/api/v1"
	"github.com/yueli-official/resource/api/internal/reserr"
	"github.com/yueli-official/resource/api/internal/resourceauthz"
)

type Authorization struct{}

func NewAuthorization() *Authorization { return &Authorization{} }

func (*Authorization) GetAuthorizationConsole(
	ctx context.Context,
	_ *v1.GetAuthorizationConsoleReq,
) (*v1.GetAuthorizationConsoleRes, error) {
	service := authorizationService(ctx)
	if service == nil {
		return nil, reserr.AuthorizationUnavailable()
	}
	revisions, err := service.Runtime().ListPolicyRevisions(ctx, authorization.PolicyRevisionListQuery{
		Actor: service.Subject(ctx), ScopeID: resourceauthz.RootScopeID, Limit: 500,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	var active, selected authorization.PolicyRevision
	for _, revision := range revisions.Revisions {
		if revision.State == authorization.PolicyActive && revision.Number > active.Number {
			active = revision
		}
		if revision.State == authorization.PolicyDraft && revision.Number > selected.Number {
			selected = revision
		}
	}
	if selected.Number == 0 {
		selected = active
	}
	snapshot, err := service.Runtime().GetPolicySnapshot(ctx, authorization.PolicySnapshotQuery{
		Actor: service.Subject(ctx), Revision: selected.Number, ScopeID: resourceauthz.RootScopeID,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	applications, err := service.Runtime().ListApplications(ctx, authorization.ApplicationListQuery{
		Actor: service.Subject(ctx), ScopeID: resourceauthz.RootScopeID,
		State: authorization.ApplicationPending, Limit: 500,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	grants, err := service.Runtime().ListGrants(ctx, authorization.GrantListQuery{
		Actor: service.Subject(ctx), ScopeID: resourceauthz.RootScopeID,
		ActiveOnly: true, Limit: 1000,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.GetAuthorizationConsoleRes{
		ActiveRevision: active.Number,
		Policy:         authorizationPolicyView(snapshot.Revision),
		Roles:          authorizationRoleViews(snapshot.Roles),
		AutomaticRules: automaticRuleViews(snapshot.AutomaticRules),
		Applications:   authorizationApplicationViews(applications.Applications),
		Grants:         authorizationGrantViews(grants.Grants),
		Capabilities:   resourceCapabilityViews(),
	}, nil
}

func (*Authorization) CreateAuthorizationDraft(
	ctx context.Context,
	req *v1.CreateAuthorizationDraftReq,
) (*v1.CreateAuthorizationDraftRes, error) {
	service := authorizationService(ctx)
	revision, err := service.Runtime().CreatePolicyDraft(ctx, authorization.CreatePolicyDraftCommand{
		Actor: service.Subject(ctx), ScopeID: resourceauthz.RootScopeID,
		ExpectedActiveRevision: req.ExpectedActiveRevision,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.CreateAuthorizationDraftRes{Policy: authorizationPolicyView(revision)}, nil
}

func (*Authorization) SetAuthorizationRoleCapabilities(
	ctx context.Context,
	req *v1.SetAuthorizationRoleCapabilitiesReq,
) (*v1.SetAuthorizationRoleCapabilitiesRes, error) {
	service := authorizationService(ctx)
	revision, err := service.Runtime().SetRoleCapabilities(ctx, authorization.SetRoleCapabilitiesCommand{
		Actor: service.Subject(ctx), Revision: req.Revision, Role: authorization.RoleKey(req.Role),
		Capabilities: capabilityKeys(req.Capabilities),
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.SetAuthorizationRoleCapabilitiesRes{Policy: authorizationPolicyView(revision)}, nil
}

func (*Authorization) CreateAuthorizationRole(
	ctx context.Context,
	req *v1.CreateAuthorizationRoleReq,
) (*v1.CreateAuthorizationRoleRes, error) {
	service := authorizationService(ctx)
	role, err := service.Runtime().CreateRole(ctx, authorization.CreateRoleCommand{
		Actor: service.Subject(ctx), Revision: req.Revision,
		Key: authorization.RoleKey(req.Key), DisplayName: req.DisplayName,
		ScopeID: resourceauthz.RootScopeID, Capabilities: capabilityKeys(req.Capabilities),
		Assignment: authorization.AssignmentPolicy{Sources: grantSources(req.Sources)},
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.CreateAuthorizationRoleRes{Role: authorizationRoleView(role)}, nil
}

func (*Authorization) RetireAuthorizationRole(
	ctx context.Context,
	req *v1.RetireAuthorizationRoleReq,
) (*v1.RetireAuthorizationRoleRes, error) {
	service := authorizationService(ctx)
	role, err := service.Runtime().RetireRole(ctx, authorization.RetireRoleCommand{
		Actor: service.Subject(ctx), Revision: req.Revision, Role: authorization.RoleKey(req.Role),
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.RetireAuthorizationRoleRes{Role: authorizationRoleView(role)}, nil
}

func (*Authorization) SetAuthorizationAutomaticRule(
	ctx context.Context,
	req *v1.SetAuthorizationAutomaticRuleReq,
) (*v1.SetAuthorizationAutomaticRuleRes, error) {
	service := authorizationService(ctx)
	revision, err := service.Runtime().SetAutomaticRuleEnabled(ctx, authorization.SetAutomaticRuleEnabledCommand{
		Actor: service.Subject(ctx), Revision: req.Revision, Rule: req.Rule, Enabled: req.Enabled,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.SetAuthorizationAutomaticRuleRes{Policy: authorizationPolicyView(revision)}, nil
}

func (*Authorization) ValidateAuthorizationPolicy(
	ctx context.Context,
	req *v1.ValidateAuthorizationPolicyReq,
) (*v1.ValidateAuthorizationPolicyRes, error) {
	service := authorizationService(ctx)
	result, err := service.Runtime().ValidatePolicy(ctx, authorization.ValidatePolicyCommand{
		Actor: service.Subject(ctx), Revision: req.Revision,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.ValidateAuthorizationPolicyRes{Valid: result.Valid, Violations: result.Violations}, nil
}

func (*Authorization) PreviewAuthorizationPolicy(
	ctx context.Context,
	req *v1.PreviewAuthorizationPolicyReq,
) (*v1.PreviewAuthorizationPolicyRes, error) {
	service := authorizationService(ctx)
	result, err := service.Runtime().PreviewPolicy(ctx, authorization.PreviewPolicyCommand{
		Actor: service.Subject(ctx), Revision: req.Revision,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.PreviewAuthorizationPolicyRes{
		AddedBindings: result.AddedBindings, RemovedBindings: result.RemovedBindings,
	}, nil
}

func (*Authorization) ActivateAuthorizationPolicy(
	ctx context.Context,
	req *v1.ActivateAuthorizationPolicyReq,
) (*v1.ActivateAuthorizationPolicyRes, error) {
	service := authorizationService(ctx)
	revision, err := service.Runtime().ActivatePolicy(ctx, authorization.ActivatePolicyCommand{
		Actor: service.Subject(ctx), Revision: req.Revision,
		ExpectedActiveRevision: req.ExpectedActiveRevision,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.ActivateAuthorizationPolicyRes{Policy: authorizationPolicyView(revision)}, nil
}

func (*Authorization) ListRequestableRoles(
	ctx context.Context,
	_ *v1.ListRequestableRolesReq,
) (*v1.ListRequestableRolesRes, error) {
	service := authorizationService(ctx)
	roles, err := service.Runtime().ListRequestableRoles(ctx, authorization.RequestableRoleQuery{
		Subject: service.Subject(ctx), ScopeID: resourceauthz.RootScopeID,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.ListRequestableRolesRes{Items: authorizationRoleViews(roles)}, nil
}

func (*Authorization) ApplyForRole(
	ctx context.Context,
	req *v1.ApplyForRoleReq,
) (*v1.ApplyForRoleRes, error) {
	service := authorizationService(ctx)
	application, err := service.Runtime().Apply(ctx, authorization.ApplyCommand{
		Actor: service.Subject(ctx), Role: authorization.RoleKey(req.Role),
		ScopeID: resourceauthz.RootScopeID, Reason: req.Reason,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.ApplyForRoleRes{Application: authorizationApplicationView(application)}, nil
}

func (*Authorization) ListMyApplications(
	ctx context.Context,
	_ *v1.ListMyApplicationsReq,
) (*v1.ListMyApplicationsRes, error) {
	service := authorizationService(ctx)
	subject := service.Subject(ctx)
	page, err := service.Runtime().ListApplications(ctx, authorization.ApplicationListQuery{
		Actor: subject, Subject: subject, ScopeID: resourceauthz.RootScopeID, Limit: 100,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.ListMyApplicationsRes{Items: authorizationApplicationViews(page.Applications)}, nil
}

func (*Authorization) WithdrawRoleApplication(
	ctx context.Context,
	req *v1.WithdrawRoleApplicationReq,
) (*v1.WithdrawRoleApplicationRes, error) {
	service := authorizationService(ctx)
	application, err := service.Runtime().WithdrawApplication(ctx, authorization.WithdrawApplicationCommand{
		Actor: service.Subject(ctx), ApplicationID: authorization.ApplicationID(req.ID),
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.WithdrawRoleApplicationRes{Application: authorizationApplicationView(application)}, nil
}

func (*Authorization) ReviewRoleApplication(
	ctx context.Context,
	req *v1.ReviewRoleApplicationReq,
) (*v1.ReviewRoleApplicationRes, error) {
	service := authorizationService(ctx)
	application, err := service.Runtime().ReviewApplication(ctx, authorization.ReviewApplicationCommand{
		Actor: service.Subject(ctx), ApplicationID: authorization.ApplicationID(req.ID),
		Decision: authorization.ReviewDecision(req.Decision), Reason: req.Reason,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.ReviewRoleApplicationRes{Application: authorizationApplicationView(application)}, nil
}

func (*Authorization) GrantAuthorizationRole(
	ctx context.Context,
	req *v1.GrantAuthorizationRoleReq,
) (*v1.GrantAuthorizationRoleRes, error) {
	service := authorizationService(ctx)
	grant, err := service.Runtime().Grant(ctx, authorization.GrantCommand{
		Actor:  service.Subject(ctx),
		Target: authorization.SubjectRef{Kind: authorization.SubjectUser, ID: req.Subject},
		Role:   authorization.RoleKey(req.Role), ScopeID: resourceauthz.RootScopeID,
		Source: authorization.GrantSourceDirect,
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.GrantAuthorizationRoleRes{Grant: authorizationGrantView(grant)}, nil
}

func (*Authorization) RevokeAuthorizationGrant(
	ctx context.Context,
	req *v1.RevokeAuthorizationGrantReq,
) (*v1.RevokeAuthorizationGrantRes, error) {
	service := authorizationService(ctx)
	grant, err := service.Runtime().Revoke(ctx, authorization.RevokeCommand{
		Actor: service.Subject(ctx), GrantID: authorization.GrantID(req.ID),
	})
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.RevokeAuthorizationGrantRes{Grant: authorizationGrantView(grant)}, nil
}

func authorizationRoleViews(roles []authorization.Role) []v1.AuthorizationRoleView {
	items := make([]v1.AuthorizationRoleView, len(roles))
	for index, role := range roles {
		items[index] = authorizationRoleView(role)
	}
	return items
}

func authorizationRoleView(role authorization.Role) v1.AuthorizationRoleView {
	capabilities := make([]string, len(role.Capabilities))
	for index, capability := range role.Capabilities {
		capabilities[index] = string(capability)
	}
	sources := make([]string, len(role.Assignment.Sources))
	for index, source := range role.Assignment.Sources {
		sources[index] = string(source)
	}
	return v1.AuthorizationRoleView{
		ID: string(role.ID), Key: string(role.Key), DisplayName: role.DisplayName,
		Kind: string(role.Kind), Status: string(role.Status), Protected: role.Protected,
		Capabilities: capabilities, Sources: sources,
	}
}

func authorizationPolicyView(revision authorization.PolicyRevision) v1.AuthorizationPolicyView {
	return v1.AuthorizationPolicyView{
		Number: revision.Number, Base: revision.Base, State: string(revision.State),
		CreatedAt: revision.CreatedAt, ActivatedAt: revision.ActivatedAt,
	}
}

func automaticRuleViews(rules []authorization.AutomaticRulePolicy) []v1.AuthorizationAutomaticRuleView {
	items := make([]v1.AuthorizationAutomaticRuleView, len(rules))
	for index, rule := range rules {
		items[index] = v1.AuthorizationAutomaticRuleView{Key: rule.Key, Enabled: rule.Enabled}
	}
	return items
}

func authorizationApplicationViews(applications []authorization.Application) []v1.AuthorizationApplicationView {
	items := make([]v1.AuthorizationApplicationView, len(applications))
	for index, application := range applications {
		items[index] = authorizationApplicationView(application)
	}
	return items
}

func authorizationApplicationView(application authorization.Application) v1.AuthorizationApplicationView {
	return v1.AuthorizationApplicationView{
		ID: string(application.ID), Subject: string(application.Subject.ID),
		Role: string(application.Role), Reason: application.Reason, State: string(application.State),
		CreatedAt: application.CreatedAt, ReviewedAt: application.ReviewedAt,
		ReviewReason: application.ReviewReason,
	}
}

func authorizationGrantViews(grants []authorization.Grant) []v1.AuthorizationGrantView {
	items := make([]v1.AuthorizationGrantView, len(grants))
	for index, grant := range grants {
		items[index] = authorizationGrantView(grant)
	}
	return items
}

func authorizationGrantView(grant authorization.Grant) v1.AuthorizationGrantView {
	return v1.AuthorizationGrantView{
		ID: string(grant.ID), Subject: grant.Target.ID,
		Role: string(grant.Role), Source: string(grant.Source),
	}
}

func capabilityKeys(values []string) []authorization.CapabilityKey {
	keys := make([]authorization.CapabilityKey, len(values))
	for index, value := range values {
		keys[index] = authorization.CapabilityKey(value)
	}
	return keys
}

func grantSources(values []string) []authorization.GrantSource {
	sources := make([]authorization.GrantSource, len(values))
	for index, value := range values {
		sources[index] = authorization.GrantSource(value)
	}
	return sources
}

func resourceCapabilityViews() []v1.AuthorizationCapabilityView {
	return []v1.AuthorizationCapabilityView{
		{Key: string(resourceauthz.CapabilityItemCreate), DisplayName: "创建资源"},
		{Key: string(resourceauthz.CapabilityItemRead), DisplayName: "查看可管理资源"},
		{Key: string(resourceauthz.CapabilityItemUpdate), DisplayName: "编辑自己的资源"},
		{Key: string(resourceauthz.CapabilityItemPublish), DisplayName: "发布自己的资源"},
		{Key: string(resourceauthz.CapabilityItemArchive), DisplayName: "归档自己的资源"},
		{Key: string(resourceauthz.CapabilityItemDelete), DisplayName: "删除自己的资源"},
		{Key: string(resourceauthz.CapabilityAssetManage), DisplayName: "管理资源文件与封面"},
		{Key: string(resourceauthz.CapabilityTagCreate), DisplayName: "创建标签"},
		{Key: string(resourceauthz.CapabilityTaxonomyManage), DisplayName: "治理分类与标签"},
		{Key: string(resourceauthz.CapabilitySiteSettingsManage), DisplayName: "管理站点设置"},
		{Key: string(resourceauthz.CapabilityAssetSettingsManage), DisplayName: "管理资源配置"},
	}
}
