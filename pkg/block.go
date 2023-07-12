package pkg

import (
	"fmt"
	"strings"

	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2/hclwrite"
	tfjson "github.com/hashicorp/terraform-json"
)

type block interface {
	address() string
	appendNewBlock(typeName string, labels []string) *hclwrite.Block
	attributes() []*attribute
	maxItems() uint64
	minItems() uint64
	nestedBlocks() []*nestedBlock
	schemaBlock() *tfjson.SchemaBlock
}

func newAttribute(parent block, name string, schema *tfjson.SchemaAttribute) *attribute {
	return &attribute{
		SchemaAttribute: schema,
		name:            name,
		parent:          parent,
	}
}

func normalizeBlockContents(b block) ([]*attribute, []*nestedBlock) {
	var attrs []*attribute
	var nbs []*nestedBlock
	linq.From(b.schemaBlock().Attributes).OrderBy(func(i interface{}) interface{} {
		return i.(linq.KeyValue).Key
	}).ForEach(func(i interface{}) {
		pair := i.(linq.KeyValue)
		name := pair.Key.(string)
		attr := newAttribute(b, name, pair.Value.(*tfjson.SchemaAttribute))
		if attr.skipAttribute() {
			return
		}
		at := attr.AttributeType
		// Some nested blocks are marked as attributes https://github.com/hashicorp/terraform-provider-azurerm/blob/v3.62.1/internal/services/containers/container_group_resource.go#L187C43-L187C43
		if at.IsObjectType() || (at.IsCollectionType() && at.ElementType().IsObjectType()) {
			nb := newNestedBlock(b, name, restoreToNestedBlockSchema(attr.SchemaAttribute))
			nbs = append(nbs, nb)
			return
		}
		attrs = append(attrs, attr)
	})
	for name, nb := range b.schemaBlock().NestedBlocks {
		nbs = append(nbs, newNestedBlock(b, name, nb))
	}
	linq.From(nbs).OrderBy(func(i interface{}) interface{} {
		return i.(*nestedBlock).name
	}).ToSlice(&nbs)
	return attrs, nbs
}

func generateVariableType(b block, rootType bool) string {
	var sb strings.Builder
	nb, isNestedBlock := b.(*nestedBlock)

	if b.maxItems() == 1 || isNestedBlock && nb.NestingMode() == tfjson.SchemaNestingModeSingle {
		sb.WriteString("object({\n")
	} else {
		collection := "set(object({\n"
		if isNestedBlock && nb.NestingMode() == tfjson.SchemaNestingModeList {
			collection = "list(object({\n"
		}
		sb.WriteString(collection)
	}

	for _, s := range b.attributes() {
		name := s.name
		attr := s.SchemaAttribute
		if s.computedOnly() || s.skipAttribute() {
			continue
		}
		attrType := ctyTypeToVariableTypeString(attr.AttributeType)
		if attr.Optional {
			sb.WriteString(fmt.Sprintf("  %s = optional(%s)\n", name, attrType))
		} else {
			sb.WriteString(fmt.Sprintf("  %s = %s\n", name, attrType))
		}
	}

	for _, s := range b.nestedBlocks() {
		name := s.name
		if !s.blockReadOnly() {
			sb.WriteString(fmt.Sprintf("  %s = %s\n", name, generateVariableType(s, false)))
		}
	}

	sb.WriteString("})")

	nb, isNestedBlock = b.(*nestedBlock)
	if isNestedBlock && b.maxItems() != 1 && (nb.NestingMode() == tfjson.SchemaNestingModeList || nb.NestingMode() == tfjson.SchemaNestingModeSet) {
		sb.WriteString(")")
	}

	t := sb.String()
	if !rootType && b.minItems() < 1 {
		t = fmt.Sprintf("optional(%s)", t)
	}
	return t
}
