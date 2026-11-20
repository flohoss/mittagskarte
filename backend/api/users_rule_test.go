package api_test

import (
	"net/http"
	"strings"
	"testing"

	_ "github.com/flohoss/mittagskarte/migrations"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tests"
)

func TestGuestCannotCreateUser(t *testing.T) {
	t.Parallel()

	scenario := tests.ApiScenario{
		Name:   "guest self signup is rejected",
		Method: http.MethodPost,
		URL:    "/api/collections/users/records",
		Body: strings.NewReader(`{
			"email": "intruder@example.com",
			"password": "secret123456",
			"passwordConfirm": "secret123456"
		}`),
		ExpectedStatus:  http.StatusForbidden,
		ExpectedContent: []string{"Only superusers can perform this action."},
	}

	scenario.Test(t)
}

func TestSuperuserCanStillCreateUser(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	t.Cleanup(app.Cleanup)

	superusers, err := app.FindCachedCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("failed to find superusers collection: %v", err)
	}

	superuser := core.NewRecord(superusers)
	superuser.SetEmail("owner@example.com")
	superuser.SetPassword("secret123456")
	if err := app.Save(superuser); err != nil {
		t.Fatalf("failed to save superuser: %v", err)
	}

	token, err := superuser.NewAuthToken()
	if err != nil {
		t.Fatalf("failed to create superuser token: %v", err)
	}

	scenario := tests.ApiScenario{
		Name:            "superuser creates a user",
		Method:          http.MethodPost,
		URL:             "/api/collections/users/records",
		Body:            strings.NewReader(`{"email":"staff@example.com","password":"secret123456","passwordConfirm":"secret123456"}`),
		Headers:         map[string]string{"Authorization": token},
		TestAppFactory:  func(t testing.TB) *tests.TestApp { return app },
		ExpectedStatus:  http.StatusOK,
		ExpectedContent: []string{"staff@example.com"},
	}

	scenario.Test(t)
}
