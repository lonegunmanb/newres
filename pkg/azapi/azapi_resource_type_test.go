package azapi_test

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestGetAzApiResourceTypeDefinition(t *testing.T) {
	loader := types.DefaultAzureSchemaLoader()
	resourceDef, err := loader.GetResourceDefinition("Microsoft.Resources/resourcegroups", "2024-07-01")
	require.NoError(t, err)
	assert.NotNil(t, resourceDef)
}

func TestAzApiTypeToTfSchemaAttribute(t *testing.T) {
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
				Flags:       []types.ObjectPropertyFlag{types.Required},
				Description: p("string field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.String,
				Description:     "string field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
				Required:        true,
			},
			description: "simple string field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.StringType{
						Type: "StringType",
					},
				},
				Flags:       []types.ObjectPropertyFlag{types.Required},
				Description: p("string field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.String,
				Description:     "string field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
				Required:        true,
			},
			description: "optional string field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.BooleanType{
						Type: "BooleanType",
					},
				},
				Flags:       []types.ObjectPropertyFlag{types.Required},
				Description: p("boolean field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Bool,
				Description:     "boolean field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
				Required:        true,
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
				Flags:       []types.ObjectPropertyFlag{types.Required},
				Description: p("number field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Number,
				Description:     "number field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
				Required:        true,
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
					Type: &types.ObjectType{
						Type: "ObjectType",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
								Flags: []types.ObjectPropertyFlag{types.Required},
							},
							"field2": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
								Flags: []types.ObjectPropertyFlag{types.ReadOnly},
							},
						},
					},
				},
				Description: p("object with readonly field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Object(map[string]cty.Type{"field1": cty.String}),
				Description:     "object with readonly field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "readonly field should be filtered out",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
								Flags: []types.ObjectPropertyFlag{types.Required},
							},
							"field2": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
								Flags: []types.ObjectPropertyFlag{types.Identifier},
							},
						},
					},
				},
				Description: p("object with identifier field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Object(map[string]cty.Type{"field1": cty.String}),
				Description:     "object with identifier field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "identifier field should be filtered out",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
								Flags: []types.ObjectPropertyFlag{types.Required},
							},
							"field2": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
							},
						},
					},
				},
				Description: p("object with optional and required field"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
					"field1": cty.String,
					"field2": cty.String,
				}, []string{"field2"}),
				Description:     "object with optional and required field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "object with optional and required field",
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
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ArrayType{
						Type: "ArrayType",
						ItemType: &types.TypeReference{
							Type: &types.StringType{
								Type: "StringType",
							},
						},
					},
				},
				Description: p("array of string"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.List(cty.String),
				Description:     "array of string",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "array of string",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ArrayType{
						Type: "ArrayType",
						ItemType: &types.TypeReference{
							Type: &types.IntegerType{
								Type: "IntegerType",
							},
						},
					},
				},
				Description: p("array of number"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.List(cty.Number),
				Description:     "array of number",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "array of number",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						AdditionalProperties: &types.TypeReference{
							Type: &types.StringType{
								Type: "StringType",
							},
						},
					},
				},
				Description: p("map of string"),
			},
			expected: &tfjson.SchemaAttribute{
				AttributeType:   cty.Map(cty.String),
				Description:     "map of string",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "map of string",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			actual := azapi.ConvertAzApiTypeToTerraformJsonSchemaAttribute(c.azApiProperty)
			require.NotNil(t, actual)
			assert.Equal(t, *c.expected, *actual)
		})
	}
}

func p[T any](value T) *T {
	return &value
}
