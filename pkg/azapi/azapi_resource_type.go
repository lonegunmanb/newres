package azapi

import (
	"fmt"
	"log"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

func GetAzApiType(resourceType, apiVersion string) (*types.ResourceType, error) {
	loader := types.DefaultAzureSchemaLoader()
	resourceDef, err := loader.GetResourceDefinition(resourceType, apiVersion)
	if err != nil {
		return nil, err
	}
	if resourceDef == nil || resourceDef.Body == nil {
		return nil, fmt.Errorf("resource %s not found", resourceType)
	}
	objectType, ok := resourceDef.Body.Type.(*types.ObjectType)
	if !ok {
		return nil, fmt.Errorf("resource %s body is not object", resourceType)
	}
	removeGeneralFields(objectType.Properties)
	return resourceDef, nil
}

func removeGeneralFields(properties map[string]types.ObjectProperty) {
}

func ConvertAzApiObjectTypeToTerraformJsonSchemaAttribute(property types.ObjectProperty) (*tfjson.SchemaBlock, error) {
	objType, ok := property.Type.Type.(*types.ObjectType)
	if !ok {
		log.Panicf("expect object type but got %v", property.Type.Type)
	}
	attributes := make(map[string]*tfjson.SchemaAttribute)
	for n, p := range objType.Properties {
		if shouldFilterOut(p) {
			continue
		}
		var schema *tfjson.SchemaAttribute
		ctyType := convertAzApiTypeToCtyType(p.Type.Type)
		if ctyType == nil {
			return nil, fmt.Errorf("unknown type %v", p.Type.Type)
		}
		schema = &tfjson.SchemaAttribute{
			AttributeType: *ctyType,
		}
		if stringType, ok := p.Type.Type.(*types.StringType); ok {
			schema.Sensitive = stringType.Sensitive
		}
		schema.Required = isRequired(p)

		if p.Description != nil {
			schema.Description = *p.Description
			schema.DescriptionKind = tfjson.SchemaDescriptionKindPlain
		}
		attributes[n] = schema
	}

	block := &tfjson.SchemaBlock{
		Attributes: attributes,
	}
	if property.Description != nil {
		block.Description = *property.Description
		block.DescriptionKind = tfjson.SchemaDescriptionKindPlain
	}
	return block, nil
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
	case *types.UnionType:
		{
			if len(t.Elements) == 0 {
				log.Panicf("empty union type: %v", t)
			}
			if _, ok := t.Elements[0].Type.(*types.StringLiteralType); ok {
				return &cty.String
			}
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
