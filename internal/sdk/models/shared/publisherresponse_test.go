package shared

import (
	"encoding/json"
	"testing"
)

// TestPublisherResponseData_ConnectedAppsAsStrings verifies that a GET-by-ID
// response where connected_apps is an array of strings (the documented API shape,
// and the shape the list endpoint always returns) unmarshals without error and
// that other fields are correctly populated.
//
// Regression test for GitHub issue #96 / BUG-017.
func TestPublisherResponseData_ConnectedAppsAsStrings(t *testing.T) {
	raw := `{
		"apps_count": 1,
		"common_name": "3418b5ec6ad1f703",
		"connected_apps": ["[BD-Gartner-HPE]"],
		"id": 10950,
		"lbrokerconnect": false,
		"name": "BD-gartner-hpe-sd-wan",
		"publisher_upgrade_profiles_id": 1,
		"registered": false,
		"status": "not registered",
		"upgrade_request": false,
		"upgrade_status": {"upstat": "not_support"}
	}`

	var data PublisherResponseData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("unmarshal failed with connected_apps as strings: %v", err)
	}

	assertPublisherFields(t, &data)
}

// TestPublisherResponseData_ConnectedAppsAsObjects verifies that a GET-by-ID
// response where connected_apps is an array of objects (the shape currently
// returned by the Netskope API, which differs from the documented string shape)
// unmarshals without error and that other fields are correctly populated.
//
// This is the shape that caused the original crash:
//
//	json: cannot unmarshal object into Go value of type string
//
// Regression test for GitHub issue #96 / BUG-017.
func TestPublisherResponseData_ConnectedAppsAsObjects(t *testing.T) {
	raw := `{
		"apps_count": 1,
		"common_name": "3418b5ec6ad1f703",
		"connected_apps": [
			{
				"access_method": "client",
				"host": "172.20.36.40",
				"last_connected": null,
				"name": "[BD-Gartner-HPE]"
			}
		],
		"id": 10950,
		"lbrokerconnect": false,
		"name": "BD-gartner-hpe-sd-wan",
		"publisher_upgrade_profiles_id": 1,
		"registered": false,
		"status": "not registered",
		"upgrade_request": false,
		"upgrade_status": {"upstat": "not_support"}
	}`

	var data PublisherResponseData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("unmarshal failed with connected_apps as objects: %v\n\nThis is the crash from GitHub issue #96. The connected_apps field must be excluded from the SDK struct (x-speakeasy-ignore: true) so the JSON decoder skips it regardless of shape.", err)
	}

	assertPublisherFields(t, &data)
}

// TestPublisherResponseData_NoConnectedApps verifies that a publisher with no
// connected apps (empty array) unmarshals correctly. This is the case that
// masked BUG-017 in testing — newly created publishers have no apps connected
// so the type mismatch was never exercised.
func TestPublisherResponseData_NoConnectedApps(t *testing.T) {
	raw := `{
		"apps_count": 0,
		"common_name": "3418b5ec6ad1f703",
		"connected_apps": [],
		"id": 10950,
		"lbrokerconnect": false,
		"name": "BD-gartner-hpe-sd-wan",
		"publisher_upgrade_profiles_id": 1,
		"registered": false,
		"status": "not registered",
		"upgrade_request": false,
		"upgrade_status": {"upstat": "not_support"}
	}`

	var data PublisherResponseData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("unmarshal failed with empty connected_apps: %v", err)
	}

	assertPublisherFields(t, &data)
}

// assertPublisherFields verifies that fields unrelated to connected_apps are
// correctly populated after unmarshaling, ensuring the fix does not regress
// parsing of other fields.
func assertPublisherFields(t *testing.T, data *PublisherResponseData) {
	t.Helper()

	if data.PublisherID == nil || *data.PublisherID != 10950 {
		t.Errorf("expected publisher_id=10950, got %v", data.PublisherID)
	}
	if data.PublisherName == nil || *data.PublisherName != "BD-gartner-hpe-sd-wan" {
		t.Errorf("expected name=%q, got %v", "BD-gartner-hpe-sd-wan", data.PublisherName)
	}
	if data.CommonName == nil || *data.CommonName != "3418b5ec6ad1f703" {
		t.Errorf("expected common_name=%q, got %v", "3418b5ec6ad1f703", data.CommonName)
	}
	if data.UpgradeStatus == nil || data.UpgradeStatus.Upstat == nil || *data.UpgradeStatus.Upstat != "not_support" {
		t.Errorf("expected upgrade_status.upstat=%q, got %v", "not_support", data.UpgradeStatus)
	}
}
