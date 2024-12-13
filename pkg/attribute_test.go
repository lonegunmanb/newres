package pkg

import (
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestRestoreNestedBlockContainsOptionalAttributeToSchemaBlockType(t *testing.T) {
	actual := restoreToNestedBlockSchema(&tfjson.SchemaAttribute{
		AttributeType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
			"foo": cty.String,
		}, []string{"foo"}),
	})
	require.Contains(t, actual.Block.Attributes, "foo")
	assert.False(t, actual.Block.Attributes["foo"].Required)
	assert.True(t, actual.Block.Attributes["foo"].Optional)
}
