package azapi_test

import (
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"testing"

	"github.com/ms-henglu/go-azure-types/types"
)

func TestGetAzApiResourceTypeDefinition(t *testing.T) {
	loader := types.DefaultAzureSchemaLoader()
	resourceDef, err := loader.GetResourceDefinition("Microsoft.App/containerApps", "2024-03-01")
	require.NoError(t, err)
	assert.NotNil(t, resourceDef)
}

func TestPrimitiveTypeToTfSchemaAttribute(t *testing.T) {
	cases := []struct {
		azApiProperty types.ObjectProperty
		expected      *tfjson.SchemaAttribute
		description   string
	}{
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.StringType{
						Type: "StringType",
					},
				},
				Description: p("string field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.String,
				Description:     "string field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "simple string field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.BooleanType{
						Type: "BooleanType",
					},
				},
				Description: p("boolean field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Bool,
				Description:     "boolean field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "simple boolean field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.IntegerType{
						Type: "IntegerType",
					},
				},
				Flags:       nil,
				Description: p("number field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Number,
				Description:     "number field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "simple number field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.IntegerType{
						Type: "IntegerType",
					},
				},
				Flags:       []types.ObjectPropertyFlag{types.Required},
				Description: p("required field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Number,
				Description:     "required field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
				Required:        true,
			},
			description: "required field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.IntegerType{
						Type: "IntegerType",
					},
				},
				Flags:       []types.ObjectPropertyFlag{types.ReadOnly},
				Description: p("readonly field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Number,
				Description:     "readonly field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
				Computed:        true,
			},
			description: "readonly field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.StringType{
						Type:      "StringType",
						Sensitive: true,
					},
				},
				Description: p("sensitive field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.String,
				Description:     "sensitive field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
				Sensitive:       true,
			},
			description: "sensitive field",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			actual := azapi.ConvertAzApiTypeToTerraformJsonSchemaAttribute(c.azApiProperty)
			assert.Equal(t, *c.expected, *actual)
		})
	}
}

func p[T any](value T) *T {
	return &value
}
