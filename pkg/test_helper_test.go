package pkg

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/require"
)

// testGetResourceSchema is a test helper that fetches schema dynamically.
func testGetResourceSchema(t *testing.T, resourceType string) *tfjson.Schema {
	t.Helper()
	schema, err := getResourceSchema(resourceType)
	require.NoError(t, err)
	return schema
}
