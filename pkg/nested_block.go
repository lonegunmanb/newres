package pkg

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

var _ block = &nestedBlock{}

type nestedBlock struct {
	*tfjson.SchemaBlockType
	name   string
	parent block
	attrs  []*attribute
	nbs    []*nestedBlock
}

func (n *nestedBlock) schemaBlock() *tfjson.SchemaBlock {
	return n.SchemaBlockType.Block
}

func (n *nestedBlock) address() string {
	return fmt.Sprintf("%s.%s", n.parent.address(), n.name)
}

func (n *nestedBlock) appendNewBlock(typeName string, labels []string) *hclwrite.Block {
	return n.parent.appendNewBlock(typeName, labels)
}

func (n *nestedBlock) blockReadOnly() bool {
	for _, attr := range n.Block.Attributes {
		if attr.Optional || attr.Required {
			return false
		}
	}

	for _, nb := range n.nbs {
		if !nb.blockReadOnly() {
			return false
		}
	}
	return true
}

func (n *nestedBlock) NestingMode() tfjson.SchemaNestingMode {
	return n.SchemaBlockType.NestingMode
}

func (n *nestedBlock) maxItems() uint64 {
	return n.SchemaBlockType.MaxItems
}

func (n *nestedBlock) minItems() uint64 {
	return n.SchemaBlockType.MinItems
}

func (n *nestedBlock) attributes() []*attribute {
	return n.attrs
}

func (n *nestedBlock) nestedBlocks() []*nestedBlock {
	return n.nbs
}

func (n *nestedBlock) writeBlock(iterator string) (*hclwrite.Block, error) {
	code := n.generateHCL(iterator)
	generatedNb, diag := hclwrite.ParseConfig([]byte(code), "", hcl.InitialPos)
	if diag.HasErrors() {
		return nil, fmt.Errorf("error when generating nested block code for %s: %s", n.address(), diag.Error())
	}
	return generatedNb.Body().Blocks()[0], nil
}

func (n *nestedBlock) generateHCL(iterator string) string {
	if n.isDynamic() {
		return n.generateDynamicBlock(iterator)
	}
	return n.generateBlockString(iterator)
}

func (n *nestedBlock) generateBlockString(iterator string) string {
	var hcl strings.Builder
	obj := fmt.Sprintf("%s.%s", iterator, n.name)
	if strings.HasPrefix(iterator, "var.") && strings.HasSuffix(iterator, n.name) {
		obj = iterator
	}
	nextIterator := fmt.Sprintf("%s.value", n.name)
	hcl.WriteString(fmt.Sprintf("dynamic \"%s\" {\n", n.name))
	forEach := fmt.Sprintf("  for_each = [%s]\n", obj)
	if (n.NestingMode() == tfjson.SchemaNestingModeSet ||
		n.NestingMode() == tfjson.SchemaNestingModeList ||
		n.NestingMode() == tfjson.SchemaNestingModeMap) && n.maxItems() != 1 {
		forEach = fmt.Sprintf("  for_each = %s\n", obj)
	}
	hcl.WriteString(forEach)
	hcl.WriteString("  content {\n")
	hcl.WriteString(n.generateHCLAttributes(nextIterator))
	hcl.WriteString(n.generateHCLNestedBlocks(nextIterator))
	hcl.WriteString("  }\n")
	hcl.WriteString("}\n")
	return hcl.String()
}

func (n *nestedBlock) generateDynamicBlock(iterator string) string {
	var hcl strings.Builder
	obj := fmt.Sprintf("%s.%s", iterator, n.name)
	if strings.HasPrefix(iterator, "var.") && strings.HasSuffix(iterator, n.name) {
		obj = iterator
	}
	hcl.WriteString(fmt.Sprintf("dynamic \"%s\" {\n", n.name))
	if n.maxItems() == 1 {
		hcl.WriteString(fmt.Sprintf("  for_each = %s == null ? [] : [%s]\n", obj, obj))
	} else if n.SchemaBlockType.NestingMode == tfjson.SchemaNestingModeMap {
		hcl.WriteString(fmt.Sprintf("  for_each = %s == null ? {} : %s\n", obj, obj))
	} else {
		hcl.WriteString(fmt.Sprintf("  for_each = %s == null ? [] : %s\n", obj, obj))
	}
	nextIterator := fmt.Sprintf("%s.value", n.name)
	hcl.WriteString("  content {\n")
	hcl.WriteString(n.generateHCLAttributes(nextIterator))
	hcl.WriteString(n.generateHCLNestedBlocks(nextIterator))
	hcl.WriteString("  }\n")
	hcl.WriteString("}\n")
	return hcl.String()
}

func appendVariableBlock(b block, variableName string, document map[string]argumentDescription) error {
	variableType := generateVariableType(b, true)
	variableType = fmt.Sprintf("type = %s", variableType)
	cfg, diag := hclwrite.ParseConfig([]byte(variableType), "", hcl.InitialPos)
	if diag.HasErrors() {
		return fmt.Errorf("incorrect parsed variable type for %s: %s, %s", b.address(), variableType, diag.Error())
	}
	vb := b.appendNewBlock("variable", []string{variableName})

	vb.Body().AppendUnstructuredTokens(cfg.BuildTokens(hclwrite.Tokens{}))
	vb.Body().AppendNewline()

	if b.minItems() == 0 {
		vb.Body().SetAttributeValue("default", cty.NullVal(cty.String))
	} else {
		vb.Body().SetAttributeValue("nullable", cty.False)
	}

	vb.Body().SetAttributeRaw("description", blockDescriptionTokens(b, document))
	return nil
}

func (n *nestedBlock) generateHCLAttributes(iterator string) string {
	var hb strings.Builder
	for _, attr := range n.attrs {
		if attr.computedOnly() {
			continue
		}
		hb.WriteString(fmt.Sprintf("    %s = %s.%s\n", attr.name, iterator, attr.name))
	}
	return hb.String()
}

func (n *nestedBlock) generateHCLNestedBlocks(iterator string) string {
	var hb strings.Builder
	for _, nb := range n.nbs {
		hb.WriteString(nb.generateHCL(iterator))
	}
	return hb.String()
}

func (n *nestedBlock) isDynamic() bool {
	switch n.SchemaBlockType.NestingMode {
	case tfjson.SchemaNestingModeList, tfjson.SchemaNestingModeSingle, tfjson.SchemaNestingModeGroup:
		return n.minItems() == 0
	case tfjson.SchemaNestingModeSet, tfjson.SchemaNestingModeMap:
		return true
	}
	panic(fmt.Sprintf("unexpected nesting mode: %s", n.SchemaBlockType.NestingMode))
}

func generateVariableDescription(n block, descriptions map[string]argumentDescription) hclwrite.Tokens {
	descriptionTokens := newTokens()
	for _, attr := range n.attributes() {
		desc := ""
		nb, isNestedBlock := n.(*nestedBlock)
		key := attr.name
		if isNestedBlock {
			key = fmt.Sprintf("%s.%s", nb.name, attr.name)
		}
		fetchedDesc, ok := descriptions[key]
		if ok {
			desc = fetchedDesc.desc
		} else {
			desc = descriptions[attr.name].desc
		}
		descriptionTokens.ident(fmt.Sprintf("- `%s` - %s", attr.name, desc), 2).newLine()
	}
	for _, nb := range n.nestedBlocks() {
		descriptionTokens.
			newLine().
			ident("---", 2).
			newLine().
			ident(fmt.Sprintf("`%s` block supports the following:", nb.name), 2).
			newLine().
			rawTokens(generateVariableDescription(nb, descriptions))
	}
	return descriptionTokens.Tokens
}
