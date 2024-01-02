package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

type attrExpr func(r *resourceBlock, name string) hclwrite.Tokens
type nestedBlockIteratorExpr func(r *resourceBlock, name string) string

var _ block = &resourceBlock{}

type resourceBlock struct {
	*tfjson.Schema
	name              string
	vendor            string
	nameWithoutVendor string
	f                 *hclwrite.File
	attrs             []*attribute
	nbs               []*nestedBlock
	writeBlock        *hclwrite.Block
	cfg               Config
}

func (r *resourceBlock) schemaBlock() *tfjson.SchemaBlock {
	return r.Block
}

func (r *resourceBlock) address() string {
	return r.nameWithoutVendor
}

func (r *resourceBlock) appendNewBlock(typeName string, labels []string) *hclwrite.Block {
	return r.f.Body().AppendNewBlock(typeName, labels)
}

func newResourceBlock(name string, schema *tfjson.Schema, cfg Config) (*resourceBlock, error) {
	resourceTypeSegments := strings.Split(name, "_")
	if len(resourceTypeSegments) < 1 {
		return nil, fmt.Errorf("incorrect resource type: %s", name)
	}
	vendor := resourceTypeSegments[0]
	nameWithoutVendor := strings.TrimPrefix(name, fmt.Sprintf("%s_", vendor))
	r := &resourceBlock{
		Schema:            schema,
		name:              name,
		vendor:            vendor,
		nameWithoutVendor: nameWithoutVendor,
		f:                 hclwrite.NewFile(),
		cfg:               cfg,
	}
	r.init()
	return r, nil
}

func (r *resourceBlock) init() {
	r.attrs, r.nbs = normalizeBlockContents(r)
	r.f = hclwrite.NewFile()
	r.writeBlock = hclwrite.NewBlock("resource", []string{r.name, "this"})
}

func (r *resourceBlock) schemaAttributeToHCLBlock(attributeName string, attribute *tfjson.SchemaAttribute, descriptions map[string]argumentDescription) *hclwrite.Block {
	name := fmt.Sprintf("%s_%s", r.nameWithoutVendor, attributeName)
	wb := hclwrite.NewBlock("variable", []string{name})

	typeStr := ctyTypeToVariableTypeString(attribute.AttributeType)
	wb.Body().SetAttributeRaw("type", hclwrite.Tokens{&hclwrite.Token{
		Type:         hclsyntax.TokenIdent,
		Bytes:        []byte(typeStr),
		SpacesBefore: 0,
	}})

	if description, ok := descriptions[attributeName]; ok {
		wb.Body().SetAttributeValue("description", cty.StringVal(description.desc))
	}
	if attribute.Sensitive {
		wb.Body().SetAttributeValue("sensitive", cty.True)
	}
	if attribute.Required {
		wb.Body().SetAttributeValue("nullable", cty.False)
	}
	if attribute.Optional {
		wb.Body().SetAttributeRaw("default", newTokens().
			ident("null", 0).Tokens)
	}

	// Set description if available
	if attribute.Description != "" {
		wb.Body().SetAttributeValue("description", cty.StringVal(attribute.Description))
	}

	return wb
}

func (r *resourceBlock) attributes() []*attribute {
	return r.attrs
}

func (r *resourceBlock) nestedBlocks() []*nestedBlock {
	return r.nbs
}

func (r *resourceBlock) maxItems() uint64 {
	return 1
}

func (r *resourceBlock) minItems() uint64 {
	return 1
}

func newNestedBlock(b block, name string, s *tfjson.SchemaBlockType) *nestedBlock {
	nb := &nestedBlock{
		SchemaBlockType: s,
		name:            name,
		parent:          b,
	}
	nb.attrs, nb.nbs = normalizeBlockContents(nb)
	return nb
}

func (r *resourceBlock) generateResource(document map[string]argumentDescription, generateVariableBlock bool, attrExpr attrExpr, nbIterator nestedBlockIteratorExpr) (string, error) {
	for _, attr := range r.attrs {
		if attr.computedOnly() {
			continue
		}
		if generateVariableBlock {
			variableBlock := r.schemaAttributeToHCLBlock(attr.name, attr.SchemaAttribute, document)
			r.appendRootBlock(variableBlock)
		}
		err := r.appendAttribute(attr.name, attrExpr)
		if err != nil {
			return "", err
		}
	}

	for _, nb := range r.nbs {
		err := r.appendNestedBlock(nb, nbIterator)
		if err != nil {
			return "", err
		}
		if !generateVariableBlock {
			continue
		}
		err = r.appendVariableBlock(nb, fmt.Sprintf("%s_%s", r.nameWithoutVendor, nb.name), document)
		if err != nil {
			return "", err
		}
	}
	r.appendRootBlock(r.writeBlock)
	return string(r.f.Bytes()), nil
}

func (r *resourceBlock) appendAttribute(name string, attrExpr attrExpr) error {
	r.setAttributeRaw(name, attrExpr(r, name))
	return nil
}

func multiVarsAttributeExpr(r *resourceBlock, name string) hclwrite.Tokens {
	return newTokens().
		ident("var", 0).
		dot().
		ident(fmt.Sprintf("%s_%s", r.nameWithoutVendor, name), 0).
		Tokens
}

func uniVarAttributeExpr(r *resourceBlock, name string) hclwrite.Tokens {
	return newTokens().
		ident("var", 0).
		dot().
		ident(r.nameWithoutVendor, 0).
		dot().
		ident(name, 0).Tokens
}

func multiVarsNestedBlockIterator(r *resourceBlock, name string) string {
	return fmt.Sprintf("var.%s_%s", r.nameWithoutVendor, name)
}

func uniVarNestedBlockIterator(r *resourceBlock, name string) string {
	return fmt.Sprintf("var.%s.%s", r.nameWithoutVendor, name)
}

func (r *resourceBlock) blockDescriptionTokens(b block, documents map[string]argumentDescription) hclwrite.Tokens {
	descriptionTokens := generateVariableDescription(b, documents)

	return newTokens().
		oHeredoc(fmt.Sprintf("<<-%s", r.cfg.GetDelimiter())).
		newLine().
		rawTokens(descriptionTokens).
		cHeredoc(r.cfg.GetDelimiter()).
		newLine().Tokens
}

func (r *resourceBlock) appendNestedBlock(nestedBlock *nestedBlock, iterator nestedBlockIteratorExpr) error {
	rootIterator := iterator(r, nestedBlock.name)
	writeBlock, err := nestedBlock.writeBlock(rootIterator)
	if err != nil {
		return err
	}
	r.writeBlock.Body().AppendBlock(writeBlock)
	r.appendNewline()
	return nil
}

func (r *resourceBlock) appendRootBlock(block *hclwrite.Block) {
	r.f.Body().AppendBlock(block)
	r.f.Body().AppendNewline()
}

func (r *resourceBlock) appendNewline() {
	r.writeBlock.Body().AppendNewline()
}

func (r *resourceBlock) setAttributeRaw(name string, tokens hclwrite.Tokens) {
	r.writeBlock.Body().SetAttributeRaw(name, tokens)
}

func (r *resourceBlock) generateUniVarResource(document map[string]argumentDescription) (string, error) {
	err := r.appendVariableBlock(r, r.nameWithoutVendor, document)
	if err != nil {
		return "", err
	}
	return r.generateResource(document, false, uniVarAttributeExpr, uniVarNestedBlockIterator)
}

func (r *resourceBlock) generateMultiVarsResource(document map[string]argumentDescription) (string, error) {
	return r.generateResource(document, true, multiVarsAttributeExpr, multiVarsNestedBlockIterator)
}

func (r *resourceBlock) appendVariableBlock(b block, variableName string, document map[string]argumentDescription) error {
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

	vb.Body().SetAttributeRaw("description", r.blockDescriptionTokens(b, document))
	return nil
}
