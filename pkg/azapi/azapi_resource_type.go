package azapi

import (
	"fmt"
	"log"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

var rootAttributes = map[string]struct{}{
	"location": {},
	"name":     {},
	"tags":     {},
	"identity": {},
}

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
	bodyAttributes, err := convertToAttributes(objType)
	if err != nil {
		return nil, err
	}

	block := &tfjson.SchemaBlock{
		Attributes: wrapBodySchema(bodyAttributes),
	}
	if property.Description != nil {
		block.Description = *property.Description
		block.DescriptionKind = tfjson.SchemaDescriptionKindPlain
	}
	return block, nil
}

func convertToAttributes(objType *types.ObjectType) (map[string]*tfjson.SchemaAttribute, error) {
	bodyAttributes := make(map[string]*tfjson.SchemaAttribute)
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
		schema.Optional = !schema.Required

		if p.Description != nil {
			schema.Description = *p.Description
			schema.DescriptionKind = tfjson.SchemaDescriptionKindPlain
		}
		bodyAttributes[n] = schema
	}
	return bodyAttributes, nil
}

func wrapBodySchema(bodyAttributes map[string]*tfjson.SchemaAttribute) map[string]*tfjson.SchemaAttribute {
	attributes := make(map[string]*tfjson.SchemaAttribute)

	for name, _ := range rootAttributes {
		if attr, ok := bodyAttributes[name]; ok {
			attributes[name] = attr
			delete(bodyAttributes, name)
		}
	}
	var optionalList []string
	var bodyTypes = make(map[string]cty.Type)
	for name, attr := range bodyAttributes {
		if attr.Optional || !attr.Required {
			optionalList = append(optionalList, name)
		}
		bodyTypes[name] = attr.AttributeType
	}
	bodyType := cty.Object(bodyTypes)
	if len(optionalList) > 0 {
		bodyType = cty.ObjectWithOptionalAttrs(bodyTypes, optionalList)
	}
	attributes["body"] = &tfjson.SchemaAttribute{
		AttributeType: bodyType,
		Required:      true,
	}
	return attributes
}

func convertAzApiTypeToCtyType(azApiType types.TypeBase) *cty.Type {
	switch t := azApiType.(type) {
	case *types.StringLiteralType:
		{
			return &cty.String
		}
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
	case *types.AnyType:
		{
			return &cty.DynamicPseudoType
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
	case *types.DiscriminatedObjectType:
		{
			return toDiscriminatedObjectType(t)
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
	default:
		{
			log.Panicf("unknown type %v", azApiType)
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

func toDiscriminatedObjectType(t *types.DiscriminatedObjectType) *cty.Type {
	return &cty.DynamicPseudoType
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
