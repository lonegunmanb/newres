package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAzApiResourceWithDiscriminatedObjectType(t *testing.T) {
	sut := azApiResourceGenerateCommand{
		resourceType: "Microsoft.RecoveryServices/vaults/backupPolicies",
		apiVersion:   "2024-10-01",
		cfg:          Config{},
	}
	schema, err := sut.Schema()
	require.NoError(t, err)
	assert.NotNil(t, schema)
	cfg, err := GenerateResource(sut)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}
