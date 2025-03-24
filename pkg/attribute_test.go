package pkg

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
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

func TestRestoreToNestedBlockSchema_DynamicPseudoType(t *testing.T) {
	tests := []struct {
		name     string
		input    *tfjson.SchemaAttribute
		expected *tfjson.SchemaBlockType
	}{
		{
			name: "Object with DynamicPseudoType field",
			input: &tfjson.SchemaAttribute{
				AttributeType: cty.Object(map[string]cty.Type{
					"field1": cty.DynamicPseudoType,
				}),
				Required: true,
			},
			expected: &tfjson.SchemaBlockType{
				NestingMode: tfjson.SchemaNestingModeSingle,
				Block: &tfjson.SchemaBlock{
					Attributes: map[string]*tfjson.SchemaAttribute{
						"field1": {
							AttributeType: cty.DynamicPseudoType,
							Required:      true,
						},
					},
					NestedBlocks: map[string]*tfjson.SchemaBlockType{},
				},

				MinItems: 1,
				MaxItems: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := restoreToNestedBlockSchema(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
