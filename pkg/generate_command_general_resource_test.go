package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoc(t *testing.T) {
	sut := generalResource{
		resourceType: "azurerm_resource_group",
	}
	docs, err := sut.Doc()
	require.NoError(t, err)
	assert.NotEmpty(t, docs)
}
