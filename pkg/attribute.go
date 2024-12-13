package pkg

import (
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

type attribute struct {
	*tfjson.SchemaAttribute
	name   string
	parent block
}

func (a *attribute) skipAttribute() bool {
	_, isResourceBlockAttribute := a.parent.(*resourceBlock)
	if a.name == "id" && isResourceBlockAttribute {
		return true
	}
	if a.computedOnly() {
		return true
	}
	return false
}

func (a *attribute) computedOnly() bool {
	return a.Computed && !a.Optional && !a.Required
}

func restoreToNestedBlockSchema(attr *tfjson.SchemaAttribute) *tfjson.SchemaBlockType {
	attributeType := attr.AttributeType
	minItems := 0
	if attr.Required {
		minItems = 1
	}
	maxItems := 0
	if attributeType.IsObjectType() {
		maxItems = 1
	}
	schemaBlock := &tfjson.SchemaBlock{
		Attributes:   map[string]*tfjson.SchemaAttribute{},
		NestedBlocks: map[string]*tfjson.SchemaBlockType{},
	}
	var fields map[string]cty.Type
	if attributeType.IsObjectType() {
		fields = attributeType.AttributeTypes()
	} else {
		fields = attributeType.ElementType().AttributeTypes()
	}
	for s, t := range fields {
		if t.IsPrimitiveType() || (t.IsCollectionType() && t.ElementType().IsPrimitiveType()) {

			newAttr := &tfjson.SchemaAttribute{
				AttributeType: t,
				Optional:      attributeType.IsObjectType() && attributeType.AttributeOptional(s),
			}
			newAttr.Required = !newAttr.Optional
			schemaBlock.Attributes[s] = newAttr
		} else {
			newNb := restoreToNestedBlockSchema(&tfjson.SchemaAttribute{
				AttributeType: t,
				Optional:      attributeType.IsObjectType() && attributeType.AttributeOptional(s),
				Required:      !attributeType.IsObjectType() || !attributeType.AttributeOptional(s),
			})
			schemaBlock.NestedBlocks[s] = newNb
		}
	}
	nb := &tfjson.SchemaBlockType{
		NestingMode: inferNestingMode(attributeType),
		Block:       schemaBlock,
		MinItems:    uint64(minItems),
		MaxItems:    uint64(maxItems),
	}
	return nb
}
