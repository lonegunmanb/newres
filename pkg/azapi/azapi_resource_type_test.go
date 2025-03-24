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
	resourceDef, err := azapi.GetAzApiType("Microsoft.Resources/resourcegroups", "2024-07-01")
	require.NoError(t, err)
	assert.NotNil(t, resourceDef)
}

func TestAzApiTypeToTfSchemaAttribute(t *testing.T) {
	cases := []struct {
		azApiProperty types.ObjectProperty
		expected      *tfjson.SchemaBlock
		description   string
	}{
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("string field"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.String}),
						Required:      true,
					},
				},
			},
			description: "simple string field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type: "StringType",
									},
								},
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("string field"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.String}),
						Required:      true,
					},
				},
			},
			description: "optional string field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": types.ObjectProperty{
								Type: &types.TypeReference{
									Type: &types.BooleanType{
										Type: "BooleanType",
									},
								},
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("boolean field"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.Bool}),
						Required:      true,
					},
				},
			},
			description: "simple boolean field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.IntegerType{
										Type: "IntegerType",
									},
								},
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("number field"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.Number}),
					},
				},
			},
			description: "simple number field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.IntegerType{
										Type: "IntegerType",
									},
								},
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("required field"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.Number}),
						Required:      true,
					},
				},
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
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.String}),
						Required:      true,
					},
				},
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
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.String}),
					},
				},
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
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{"field1": cty.String, "field2": cty.String}, []string{"field2"}),
					},
				},
				Description:     "object with optional and required field",
				DescriptionKind: tfjson.SchemaDescriptionKindPlain,
			},
			description: "object with optional and required field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.StringType{
										Type:      "StringType",
										Sensitive: true,
									},
								},
								Description: p("sensitive field"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						AttributeType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{"field1": cty.String}, []string{"field1"}),
						Required:      true,
					},
				},
			},
			description: "sensitive field",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
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
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("array of string"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.List(cty.String)}),
					},
				},
			},
			description: "array of string",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
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
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("array of number"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.List(cty.Number)}),
					},
				},
			},
			description: "array of number",
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
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
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("map of string"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.Map(cty.String)}),
					},
				},
			},
			description: "map of string",
		},
		{
			description: "union string type",
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.UnionType{
										Type: "UnionType",
										Elements: []*types.TypeReference{
											{
												Type: &types.StringLiteralType{
													Type:  "StringLiteralType",
													Value: "value1",
												},
											},
											{
												Type: &types.StringLiteralType{
													Type:  "StringLiteralType",
													Value: "value2",
												},
											},
										},
									},
								},
								Flags: []types.ObjectPropertyFlag{types.Required},
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.String}),
					},
				},
			},
		},
		{
			azApiProperty: types.ObjectProperty{
				Type: &types.TypeReference{
					Type: &types.ObjectType{
						Type: "ObjectType",
						Name: "obj",
						Properties: map[string]types.ObjectProperty{
							"field1": {
								Type: &types.TypeReference{
									Type: &types.AnyType{
										Type: "AnyType",
									},
								},
								Flags:       []types.ObjectPropertyFlag{types.Required},
								Description: p("any type field"),
							},
						},
					},
				},
			},
			expected: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					"body": {
						Required:      true,
						AttributeType: cty.Object(map[string]cty.Type{"field1": cty.DynamicPseudoType}),
					},
				},
			},
			description: "any type field",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			actual, err := azapi.ConvertAzApiObjectTypeToTerraformJsonSchemaAttribute(c.azApiProperty)
			require.NoError(t, err)
			require.NotNil(t, actual)
			assert.Equal(t, *c.expected, *actual)
		})
	}
}

func p[T any](value T) *T {
	return &value
}
