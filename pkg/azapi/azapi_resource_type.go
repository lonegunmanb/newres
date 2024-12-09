package azapi

import (
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

func ConvertAzApiTypeToTerraformJsonSchemaAttribute(property types.ObjectProperty) *tfjson.SchemaAttribute {
	var schema *tfjson.SchemaAttribute
	switch t := property.Type.Type.(type) {
	case *types.StringType:
		{
			schema = &tfjson.SchemaAttribute{
				AttributeType: cty.String,
				Sensitive:     t.Sensitive,
			}
		}
	case *types.IntegerType:
		{
			schema = &tfjson.SchemaAttribute{
				AttributeType: cty.Number,
			}
		}
	case *types.BooleanType:
		{
			schema = &tfjson.SchemaAttribute{
				AttributeType: cty.Bool,
			}
		}
	}
	if schema == nil {
		return nil
	}
	for _, flag := range property.Flags {
		if flag == types.Required {
			schema.Required = true
		}
		if flag == types.ReadOnly {
			schema.Computed = true
		}
	}
	if property.Description != nil {
		schema.Description = *property.Description
		schema.DescriptionKind = tfjson.SchemaDescriptionKindPlain
	}
	return schema
}
