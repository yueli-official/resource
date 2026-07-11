package controller

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"platform/gokit/authjwt"
)

func TestIsAdminUsesSiteOperatorsNotGlobalRole(t *testing.T) {
	adapter, err := gcfg.NewAdapterContent("resource:\n  operatorSubs: [site-operator]\n")
	if err != nil {
		t.Fatal(err)
	}
	g.Cfg().SetAdapter(adapter)

	operator := authjwt.WithPrincipal(context.Background(), &authjwt.Principal{Subject: "site-operator"})
	if !isAdmin(operator) {
		t.Fatal("site operator rejected")
	}

	globalAdmin := authjwt.WithPrincipal(context.Background(), &authjwt.Principal{Subject: "other", Roles: []string{"admin"}})
	if isAdmin(globalAdmin) {
		t.Fatal("global admin unexpectedly became resource operator")
	}
}
