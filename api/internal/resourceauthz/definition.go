// Package resourceauthz owns Resource's instance-local authorization declaration.
// Foundation owns execution and persistence; Resource owns capabilities, scopes,
// roles, predicates, and resource relations.
package resourceauthz

import (
	"context"

	"github.com/yueli-official/foundation/go/authorization"
)

const (
	RootScopeID authorization.ScopeID = "resource"

	ScopeSite     authorization.ScopeType = "site"
	ScopeResource authorization.ScopeType = "resource"

	RoleAdministrator authorization.RoleKey = "administrator"
	RoleContributor   authorization.RoleKey = "contributor"

	CapabilityPublicRead          authorization.CapabilityKey = "resource.public.read"
	CapabilityItemCreate          authorization.CapabilityKey = "resource.item.create"
	CapabilityItemRead            authorization.CapabilityKey = "resource.item.read"
	CapabilityItemUpdate          authorization.CapabilityKey = "resource.item.update"
	CapabilityItemPublish         authorization.CapabilityKey = "resource.item.publish"
	CapabilityItemArchive         authorization.CapabilityKey = "resource.item.archive"
	CapabilityItemDelete          authorization.CapabilityKey = "resource.item.delete"
	CapabilityAssetManage         authorization.CapabilityKey = "resource.asset.manage"
	CapabilityTagCreate           authorization.CapabilityKey = "resource.tag.create"
	CapabilityTaxonomyManage      authorization.CapabilityKey = "resource.taxonomy.manage"
	CapabilitySiteSettingsManage  authorization.CapabilityKey = "resource.site_settings.manage"
	CapabilityAssetSettingsManage authorization.CapabilityKey = "resource.asset_settings.manage"

	RelationOwner authorization.RelationKind = "owner"

	ConstraintNormalRoleOwnsResource    authorization.ConstraintKey = "resource.normal_role_owns_resource"
	PredicateRegistrationContributor    authorization.PredicateKey  = "resource.registration_auto_contributor"
	TriggerUserRegistered               authorization.TriggerKey    = "identity.user.registered"
	AutomaticRegistrationContributorKey                             = "resource.registration_contributor"
)

func ResourceScopeID(id string) authorization.ScopeID {
	return authorization.ScopeID("resource:" + id)
}

func Definition() authorization.Definition {
	return authorization.Definition{
		Consumer: "resource",
		Version:  1,
		Capabilities: []authorization.CapabilityDefinition{
			{
				Key: CapabilityPublicRead, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite, ScopeResource},
			},
			normalCapability(CapabilityItemCreate, ScopeSite),
			ownedCapability(CapabilityItemRead, ScopeSite, ScopeResource),
			ownedCapability(CapabilityItemUpdate, ScopeResource),
			ownedCapability(CapabilityItemPublish, ScopeResource),
			ownedCapability(CapabilityItemArchive, ScopeResource),
			ownedCapability(CapabilityItemDelete, ScopeResource),
			ownedCapability(CapabilityAssetManage, ScopeResource),
			normalCapability(CapabilityTagCreate, ScopeSite),
			protectedCapability(CapabilityTaxonomyManage),
			protectedCapability(CapabilitySiteSettingsManage),
			protectedCapability(CapabilityAssetSettingsManage),
		},
		Scopes: authorization.ScopeSchema{Types: []authorization.ScopeTypeDefinition{
			{Key: ScopeSite, Root: true, Children: []authorization.ScopeType{ScopeResource}},
			{Key: ScopeResource},
		}},
		AccessLayers: []authorization.AccessLayerDefinition{
			{
				Key:          authorization.AccessLayerVisitor,
				Capabilities: []authorization.CapabilityKey{CapabilityPublicRead},
			},
			{
				Key: authorization.AccessLayerAuthenticated,
				Capabilities: []authorization.CapabilityKey{
					authorization.CapabilityApplicationCreate,
					authorization.CapabilityApplicationReadOwn,
					authorization.CapabilityApplicationWithdraw,
					authorization.CapabilityInvitationAccept,
				},
			},
		},
		Roles: []authorization.RoleDefinition{
			{
				Key: RoleAdministrator, DisplayName: "管理员", Protected: true,
				Capabilities: []authorization.CapabilityKey{
					authorization.CapabilityManage,
					authorization.CapabilityAuditRead,
					CapabilityItemCreate,
					CapabilityItemRead,
					CapabilityItemUpdate,
					CapabilityItemPublish,
					CapabilityItemArchive,
					CapabilityItemDelete,
					CapabilityAssetManage,
					CapabilityTagCreate,
					CapabilityTaxonomyManage,
					CapabilitySiteSettingsManage,
					CapabilityAssetSettingsManage,
				},
			},
			{
				Key: RoleContributor, DisplayName: "贡献者",
				Capabilities: []authorization.CapabilityKey{
					CapabilityItemCreate,
					CapabilityItemRead,
					CapabilityItemUpdate,
					CapabilityItemPublish,
					CapabilityItemArchive,
					CapabilityItemDelete,
					CapabilityAssetManage,
					CapabilityTagCreate,
				},
				Assignment: authorization.AssignmentPolicy{Sources: []authorization.GrantSource{
					authorization.GrantSourceApplication,
					authorization.GrantSourceInvitation,
					authorization.GrantSourceDirect,
					authorization.GrantSourceAutomatic,
					authorization.GrantSourceGroup,
				}},
			},
		},
		Constraints: []authorization.ConstraintDefinition{{
			Key: ConstraintNormalRoleOwnsResource, Version: 1,
			Mode: authorization.ConstraintSource,
			Capabilities: []authorization.CapabilityKey{
				CapabilityItemRead,
				CapabilityItemUpdate,
				CapabilityItemPublish,
				CapabilityItemArchive,
				CapabilityItemDelete,
				CapabilityAssetManage,
			},
			AllNormalRoles: true,
		}},
		Automatic: []authorization.AutomaticRuleDefinition{{
			Key: AutomaticRegistrationContributorKey, Trigger: TriggerUserRegistered,
			Predicate: PredicateRegistrationContributor, Role: RoleContributor, Enabled: true,
		}},
	}
}

func ConstraintEvaluators() map[authorization.ConstraintKey]authorization.ConstraintEvaluator {
	return map[authorization.ConstraintKey]authorization.ConstraintEvaluator{
		ConstraintNormalRoleOwnsResource: authorization.ConstraintFunc(
			func(_ context.Context, input authorization.ConstraintInput) authorization.ConstraintResult {
				for _, owner := range input.Resource.Relations[RelationOwner] {
					if owner == input.Subject {
						return authorization.ConstraintResult{}
					}
				}
				return authorization.ConstraintResult{Denied: true}
			},
		),
	}
}

func PredicateEvaluators() map[authorization.PredicateKey]authorization.PredicateEvaluator {
	return map[authorization.PredicateKey]authorization.PredicateEvaluator{
		PredicateRegistrationContributor: authorization.PredicateFunc(
			func(_ context.Context, input authorization.PredicateInput) bool {
				return input.Subject.Kind == authorization.SubjectUser
			},
		),
	}
}

func normalCapability(
	key authorization.CapabilityKey,
	scopes ...authorization.ScopeType,
) authorization.CapabilityDefinition {
	return authorization.CapabilityDefinition{
		Key: key, Version: 1, Binding: authorization.BindingNormal,
		AllowedScopes: scopes,
		EligibleSubjects: []authorization.SubjectKind{
			authorization.SubjectUser,
			authorization.SubjectService,
		},
		Delegable: true,
	}
}

func ownedCapability(
	key authorization.CapabilityKey,
	scopes ...authorization.ScopeType,
) authorization.CapabilityDefinition {
	definition := normalCapability(key, scopes...)
	definition.QueryableRelation = RelationOwner
	return definition
}

func protectedCapability(key authorization.CapabilityKey) authorization.CapabilityDefinition {
	return authorization.CapabilityDefinition{
		Key: key, Version: 1, Binding: authorization.BindingProtectedOnly,
		Risk: authorization.RiskHigh, Audit: authorization.AuditFull,
		AllowedScopes: []authorization.ScopeType{ScopeSite},
		EligibleSubjects: []authorization.SubjectKind{
			authorization.SubjectUser,
			authorization.SubjectService,
		},
	}
}
