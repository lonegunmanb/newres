package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/matt-FFFFFF/tfpluginschema"
)

var (
	schemaServer     *tfpluginschema.Server
	schemaServerOnce sync.Once
)

func getSchemaServer() *tfpluginschema.Server {
	schemaServerOnce.Do(func() {
		schemaServer = tfpluginschema.NewServer(nil)
	})
	return schemaServer
}

// CleanupSchemaServer releases resources held by the schema server.
// Should be called when the application exits.
func CleanupSchemaServer() {
	if schemaServer != nil {
		schemaServer.Cleanup()
	}
}

// getResourceSchema dynamically retrieves the Terraform resource schema
// for the given resource type by downloading the provider binary via the
// OpenTofu registry and querying it over gRPC.
// If namespace is empty, it falls back to a default based on the provider type.
// If version is empty, the latest version is fetched from the Terraform Registry.
func getResourceSchema(resourceType string, namespace string, version string) (*tfjson.Schema, error) {
	if !resourceTypeValid(resourceType) {
		return nil, fmt.Errorf("invalid resource type: %s", resourceType)
	}
	providerType := resourceVendor(resourceType)
	if namespace == "" {
		namespace = defaultNamespace(providerType)
	}
	if version == "" {
		var err error
		version, err = getLatestProviderVersion(namespace, providerType)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version for provider %s/%s: %w", namespace, providerType, err)
		}
	} else {
		version = strings.TrimPrefix(version, "v")
	}

	req := tfpluginschema.Request{
		Namespace: namespace,
		Name:      providerType,
		Version:   version,
	}
	server := getSchemaServer()
	schema, err := server.GetResourceSchema(req, resourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource schema for %s: %w", resourceType, err)
	}
	return schema, nil
}

func defaultNamespace(providerType string) string {
	switch providerType {
	case "azapi":
		return "Azure"
	case "msgraph":
		return "microsoft"
	case "alicloud":
		return "aliyun"
	default:
		return "hashicorp"
	}
}

func getLatestProviderVersion(namespace, providerType string) (string, error) {
	url := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/%s", namespace, providerType)

	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return "", fmt.Errorf("failed to fetch provider info from registry: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry API returned status %d for provider %s/%s", resp.StatusCode, namespace, providerType)
	}

	var providerInfo struct {
		Tag string `json:"tag"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&providerInfo); err != nil {
		return "", fmt.Errorf("failed to decode provider info response: %w", err)
	}

	if providerInfo.Tag == "" {
		return "", fmt.Errorf("no tag found in provider info for %s/%s", namespace, providerType)
	}

	return strings.TrimPrefix(providerInfo.Tag, "v"), nil
}
