package admin

import "testing"

func TestParseKiroImportEntriesSupportsKiroManageExport(t *testing.T) {
	req := KiroImportRequest{
		Data: `{
			"version": 1,
			"exportedAt": "2026-06-02T01:00:00.000Z",
			"accounts": [
				{
					"email": "user@example.com",
					"password": "example-password",
					"idp": "BuilderId",
					"status": "active",
					"credentials": {
						"refreshToken": "refresh-token",
						"clientId": "client-id",
						"clientSecret": "client-secret",
						"accessToken": "access-token",
						"csrfToken": "csrf-token",
						"region": "us-east-1",
						"authMethod": "IdC",
						"provider": "BuilderId",
						"expiresAt": "1780368705542"
					},
					"subscription": {"type": "Free", "title": "KIRO FREE"},
					"usage": {"current": 0, "limit": 50},
					"tags": "batch-1",
					"lastUsedAt": "1780365105542",
					"id": "kiro-manage-id",
					"machineId": "machine-id",
					"createdAt": "1780365105542",
					"userId": "user-id",
					"lastCheckedAt": "1780368155425"
				}
			]
		}`,
		DefaultLoginType: "builder",
	}

	entries, err := parseKiroImportEntries(req)
	if err != nil {
		t.Fatalf("parseKiroImportEntries error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	item := entries[0].Item
	if item.Email != "user@example.com" {
		t.Fatalf("Email = %q, want user@example.com", item.Email)
	}
	if item.RefreshToken != "refresh-token" {
		t.Fatalf("RefreshToken = %q, want refresh-token", item.RefreshToken)
	}
	if item.ClientID != "client-id" {
		t.Fatalf("ClientID = %q, want client-id", item.ClientID)
	}
	if item.ClientSecret != "client-secret" {
		t.Fatalf("ClientSecret = %q, want client-secret", item.ClientSecret)
	}
	if item.AccessToken != "access-token" {
		t.Fatalf("AccessToken = %q, want access-token", item.AccessToken)
	}
	if item.Region != "us-east-1" {
		t.Fatalf("Region = %q, want us-east-1", item.Region)
	}
	if item.LoginType != "idc" {
		t.Fatalf("LoginType = %q, want idc", item.LoginType)
	}
	if item.MachineID != "machine-id" {
		t.Fatalf("MachineID = %q, want machine-id", item.MachineID)
	}

	credentials := buildKiroImportCredentials(item)
	for key, want := range map[string]string{
		"email":         "user@example.com",
		"access_token":  "access-token",
		"refresh_token": "refresh-token",
		"client_id":     "client-id",
		"client_secret": "client-secret",
		"region":        "us-east-1",
		"login_type":    "idc",
		"machine_id":    "machine-id",
	} {
		if got := credentials[key]; got != want {
			t.Fatalf("credentials[%s] = %v, want %s", key, got, want)
		}
	}

	if item.Extra == nil {
		t.Fatal("Extra is nil, want kiro_manage metadata")
	}
	manage, ok := item.Extra["kiro_manage"].(map[string]any)
	if !ok {
		t.Fatalf("Extra[kiro_manage] = %T, want map[string]any", item.Extra["kiro_manage"])
	}
	for key, want := range map[string]string{
		"id":                     "kiro-manage-id",
		"user_id":                "user-id",
		"idp":                    "BuilderId",
		"status":                 "active",
		"tags":                   "batch-1",
		"created_at":             "1780365105542",
		"last_used_at":           "1780365105542",
		"last_checked_at":        "1780368155425",
		"machine_id":             "machine-id",
		"auth_method":            "IdC",
		"provider":               "BuilderId",
		"csrf_token":             "csrf-token",
		"credentials_expires_at": "1780368705542",
	} {
		if got := manage[key]; got != want {
			t.Fatalf("kiro_manage[%s] = %v, want %s", key, got, want)
		}
	}
	if _, ok := manage["subscription"]; !ok {
		t.Fatal("kiro_manage.subscription missing")
	}
	if _, ok := manage["usage"]; !ok {
		t.Fatal("kiro_manage.usage missing")
	}
}
