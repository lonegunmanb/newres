package azapi

import (
	"log"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

func ConvertAzApiTypeToTerraformJsonSchemaAttribute(property types.ObjectProperty) *tfjson.SchemaAttribute {
	var schema *tfjson.SchemaAttribute
	ctyType := convertAzApiTypeToCtyType(property.Type.Type)
	if ctyType != nil {
		schema = &tfjson.SchemaAttribute{
			AttributeType: *ctyType,
		}
		if stringType, ok := property.Type.Type.(*types.StringType); ok {
			schema.Sensitive = stringType.Sensitive
		}
	}

	if schema == nil {
		return nil
	}
	for _, flag := range property.Flags {
		if flag == types.Required {
			schema.Required = true
		}
	}
	if property.Description != nil {
		schema.Description = *property.Description
		schema.DescriptionKind = tfjson.SchemaDescriptionKindPlain
	}
	return schema
}

func convertAzApiTypeToCtyType(azApiType types.TypeBase) *cty.Type {
	switch t := azApiType.(type) {
	case *types.StringType:
		{
			return &cty.String
		}
	case *types.IntegerType:
		{
			return &cty.Number
		}
	case *types.BooleanType:
		{
			return &cty.Bool
		}
	case *types.ArrayType:
		{
			return toArrayType(t)
		}
	case *types.ObjectType:
		{
			if len(t.Properties) == 0 && t.AdditionalProperties != nil {
				return toMapType(t)
			}
			return toObjectType(t)
		}
	}
	return nil
}

func toObjectType(t *types.ObjectType) *cty.Type {
	properties := make(map[string]cty.Type)
	var optionalList []string

	for n, p := range t.Properties {
		if shouldFilterOut(p) {
			continue
		}
		if !isRequired(p) {
			optionalList = append(optionalList, n)
		}
		ctyType := convertAzApiTypeToCtyType(p.Type.Type)
		if ctyType == nil {
			log.Panicf("unknown type %v", p.Type.Type)
		}
		properties[n] = *ctyType
	}
	ctyType := cty.Object(properties)
	if len(optionalList) > 0 {
		ctyType = cty.ObjectWithOptionalAttrs(properties, optionalList)
	}
	return &ctyType
}

func isRequired(p types.ObjectProperty) bool {
	for _, flag := range p.Flags {
		if flag == types.Required {
			return true
		}
	}
	return false
}

func shouldFilterOut(p types.ObjectProperty) bool {
	for _, flag := range p.Flags {
		if flag == types.ReadOnly || flag == types.Identifier {
			return true
		}
	}
	return false
}

func toArrayType(t *types.ArrayType) *cty.Type {
	itemType := convertAzApiTypeToCtyType(t.ItemType.Type)
	if itemType == nil {
		return nil
	}
	ctyType := cty.List(*itemType)
	return &ctyType
}

func toMapType(t *types.ObjectType) *cty.Type {
	elementType := convertAzApiTypeToCtyType(t.AdditionalProperties.Type)
	if elementType == nil {
		return nil
	}
	ctyType := cty.Map(*elementType)
	return &ctyType
}
