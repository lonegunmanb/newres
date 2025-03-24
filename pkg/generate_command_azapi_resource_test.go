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

func TestAzApiResourceWithAnyType(t *testing.T) {
	sut := azApiResourceGenerateCommand{
		resourceType: "Microsoft.ContainerService/managedClusters",
		apiVersion:   "2024-10-02-preview",
		cfg:          Config{},
	}
	schema, err := sut.Schema()
	require.NoError(t, err)
	assert.NotNil(t, schema)
	cfg, err := GenerateResource(sut)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestAzApiResourceWithStringLiteralType(t *testing.T) {
	sut := azApiResourceGenerateCommand{
		resourceType: "Microsoft.DocumentDB/databaseAccounts",
		apiVersion:   "2024-08-15",
		cfg:          Config{},
	}
	schema, err := sut.Schema()
	require.NoError(t, err)
	assert.NotNil(t, schema)
	cfg, err := GenerateResource(sut)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}
