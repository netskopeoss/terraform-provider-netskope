// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/operations"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/shared"
)

func init() {
	resource.AddTestSweepers("netskope_npa_private_app", &resource.Sweeper{
		Name: "netskope_npa_private_app",
		F:    sweepNPAPrivateApps,
		// Private apps must be deleted before publishers they reference
		Dependencies: []string{},
	})

	resource.AddTestSweepers("netskope_npa_publisher", &resource.Sweeper{
		Name: "netskope_npa_publisher",
		F:    sweepNPAPublishers,
		// Publishers should be deleted after private apps
		Dependencies: []string{"netskope_npa_private_app"},
	})

	resource.AddTestSweepers("netskope_npa_policy_groups", &resource.Sweeper{
		Name: "netskope_npa_policy_groups",
		F:    sweepNPAPolicyGroups,
		// Policy groups should be deleted after rules
		Dependencies: []string{"netskope_npa_rules"},
	})

	resource.AddTestSweepers("netskope_npa_rules", &resource.Sweeper{
		Name: "netskope_npa_rules",
		F:    sweepNPARules,
		// Rules should be deleted first (they reference apps, groups)
		Dependencies: []string{},
	})

	resource.AddTestSweepers("netskope_rbac_label", &resource.Sweeper{
		Name: "netskope_rbac_label",
		F:    sweepRBACLabels,
		Dependencies: []string{},
	})

	resource.AddTestSweepers("netskope_device_classification_tag", &resource.Sweeper{
		Name: "netskope_device_classification_tag",
		F:    sweepDeviceClassificationTags,
		Dependencies: []string{},
	})
}

// TestMain runs the sweepers if the -sweep flag is passed
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// getSweepClient creates an SDK client for sweep operations using environment variables
func getSweepClient() (*sdk.TerraformProviderNs, error) {
	apiKey := os.Getenv("NETSKOPE_API_KEY")
	serverURL := os.Getenv("NETSKOPE_SERVER_URL")

	if apiKey == "" || serverURL == "" {
		return nil, fmt.Errorf("NETSKOPE_API_KEY and NETSKOPE_SERVER_URL must be set for sweeping")
	}

	client := sdk.New(
		sdk.WithServerURL(serverURL),
		sdk.WithSecurity(shared.Security{
			APIKey: apiKey,
		}),
	)

	return client, nil
}

// sweepNPAPrivateApps removes all NPA Private Apps with the test prefix
func sweepNPAPrivateApps(region string) error {
	client, err := getSweepClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// List all private apps
	res, err := client.ListNPAPrivateApps(ctx, operations.ListNPAPrivateAppsRequest{})
	if err != nil {
		return fmt.Errorf("error listing private apps for sweeping: %w", err)
	}

	if res.PrivateAppsListResponse == nil || res.PrivateAppsListResponse.Data == nil {
		return nil
	}

	for _, app := range res.PrivateAppsListResponse.Data.PrivateApps {
		// Use Name field which contains the app name
		if app.Name == nil {
			continue
		}

		if !strings.HasPrefix(*app.Name, testAccResourcePrefix) {
			continue
		}

		if app.PrivateAppID == nil {
			continue
		}

		log.Printf("[INFO] Sweeping NPA Private App: %s (ID: %d)", *app.Name, *app.PrivateAppID)

		_, err := client.DeleteNPAPrivateApp(ctx, operations.DeleteNPAPrivateAppRequest{
			PrivateAppID: *app.PrivateAppID,
		})
		if err != nil {
			log.Printf("[WARN] Error deleting private app %s: %s", *app.Name, err)
		}
	}

	return nil
}

// sweepNPAPublishers removes all NPA Publishers with the test prefix
func sweepNPAPublishers(region string) error {
	client, err := getSweepClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// List all publishers
	res, err := client.NPAPublishers.ListObjects(ctx)
	if err != nil {
		return fmt.Errorf("error listing publishers for sweeping: %w", err)
	}

	if res.PublishersGetResponse == nil || res.PublishersGetResponse.Data == nil {
		return nil
	}

	for _, pub := range res.PublishersGetResponse.Data.Publishers {
		if pub.PublisherName == nil {
			continue
		}

		if !strings.HasPrefix(*pub.PublisherName, testAccResourcePrefix) {
			continue
		}

		if pub.PublisherID == nil {
			continue
		}

		log.Printf("[INFO] Sweeping NPA Publisher: %s (ID: %d)", *pub.PublisherName, *pub.PublisherID)

		_, err := client.NPAPublisher.Delete(ctx, operations.DeleteNPAPublishersRequest{
			PublisherID: *pub.PublisherID,
		})
		if err != nil {
			log.Printf("[WARN] Error deleting publisher %s: %s", *pub.PublisherName, err)
		}
	}

	return nil
}

// sweepNPAPolicyGroups removes all NPA Policy Groups with the test prefix
func sweepNPAPolicyGroups(region string) error {
	client, err := getSweepClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// List all policy groups
	res, err := client.ListNPAPolicyGroups(ctx, operations.ListNPAPolicyGroupsRequest{})
	if err != nil {
		return fmt.Errorf("error listing policy groups for sweeping: %w", err)
	}

	if res.NpaPolicygroupResponse == nil {
		return nil
	}

	for _, group := range res.NpaPolicygroupResponse.Data {
		if group.GroupName == nil {
			continue
		}

		if !strings.HasPrefix(*group.GroupName, testAccResourcePrefix) {
			continue
		}

		if group.ID == nil {
			continue
		}

		log.Printf("[INFO] Sweeping NPA Policy Group: %s (ID: %s)", *group.GroupName, *group.ID)

		_, err := client.DeleteNPAPolicyGroups(ctx, operations.DeleteNPAPolicyGroupsRequest{
			ID: *group.ID,
		})
		if err != nil {
			log.Printf("[WARN] Error deleting policy group %s: %s", *group.GroupName, err)
		}
	}

	return nil
}

// sweepNPARules removes all NPA Rules with the test prefix
func sweepNPARules(region string) error {
	client, err := getSweepClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// List all rules
	res, err := client.ListNPARules(ctx, operations.ListNPARulesRequest{})
	if err != nil {
		return fmt.Errorf("error listing rules for sweeping: %w", err)
	}

	if res.NpaPolicyResponse == nil {
		return nil
	}

	for _, rule := range res.NpaPolicyResponse.Data {
		if rule.RuleName == nil {
			continue
		}

		if !strings.HasPrefix(*rule.RuleName, testAccResourcePrefix) {
			continue
		}

		if rule.ID == nil {
			continue
		}

		log.Printf("[INFO] Sweeping NPA Rule: %s (ID: %s)", *rule.RuleName, *rule.ID)

		_, err := client.DeleteNPARules(ctx, operations.DeleteNPARulesRequest{
			ID: *rule.ID,
		})
		if err != nil {
			log.Printf("[WARN] Error deleting rule %s: %s", *rule.RuleName, err)
		}
	}

	return nil
}

// sweepRBACLabels removes all RBAC labels with the test prefix.
// Child labels must be deleted before parents, so we delete labels with parentId first.
func sweepRBACLabels(region string) error {
	client, err := getSweepClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	res, err := client.Rbac.ListRBACLabels(ctx, operations.ListRBACLabelsRequest{})
	if err != nil {
		return fmt.Errorf("error listing RBAC labels for sweeping: %w", err)
	}

	if res.RbacLabelListResponse == nil {
		return nil
	}

	// Collect labels to delete, children first (those with parentId), then parents
	var children, parents []struct {
		id   string
		name string
	}

	for _, label := range res.RbacLabelListResponse.Labels {
		if label.Name == nil || label.LabelID == nil {
			continue
		}

		if !strings.HasPrefix(*label.Name, testAccResourcePrefix) {
			continue
		}

		entry := struct {
			id   string
			name string
		}{id: *label.LabelID, name: *label.Name}

		if label.ParentID != nil && *label.ParentID != "" {
			children = append(children, entry)
		} else {
			parents = append(parents, entry)
		}
	}

	// Delete children first, then parents
	for _, label := range append(children, parents...) {
		log.Printf("[INFO] Sweeping RBAC Label: %s (ID: %s)", label.name, label.id)

		_, err := client.Rbac.DeleteRBACLabel(ctx, operations.DeleteRBACLabelRequest{
			LabelID: label.id,
		})
		if err != nil {
			log.Printf("[WARN] Error deleting RBAC label %s: %s", label.name, err)
		}
	}

	return nil
}

func sweepDeviceClassificationTags(region string) error {
	client, err := getSweepClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	res, err := client.Deviceclassification.ListDeviceClassificationTags(ctx)
	if err != nil {
		return fmt.Errorf("error listing device classification tags for sweeping: %w", err)
	}

	if res.DeviceclassificationTagListResponse == nil {
		return nil
	}

	for _, tag := range res.DeviceclassificationTagListResponse.Tags {
		if tag.Name == nil || tag.TagID == nil {
			continue
		}

		if !strings.HasPrefix(*tag.Name, testAccResourcePrefix) {
			continue
		}

		log.Printf("[INFO] Sweeping Device Classification Tag: %s (ID: %d)", *tag.Name, *tag.TagID)

		_, err := client.Deviceclassification.DeleteDeviceClassificationTag(ctx, operations.DeleteDeviceClassificationTagRequest{
			TagID: *tag.TagID,
		})
		if err != nil {
			log.Printf("[WARN] Error deleting device classification tag %s: %s", *tag.Name, err)
		}
	}

	return nil
}
