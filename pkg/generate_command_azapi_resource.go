package pkg

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

var _ ResourceGenerateCommand = azApiResourceGenerateCommand{}
var _ postProcessor = azApiResourceGenerateCommand{}

type azApiResourceGenerateCommand struct {
	resourceType string
	apiVersion   string
	cfg          Config
}

func (a azApiResourceGenerateCommand) action(terraformConfig string, cfg Config) (string, error) {
	hclCfg, diag := hclwrite.ParseConfig([]byte(terraformConfig), "", hcl.InitialPos)
	if diag.HasErrors() {
		return "", diag
	}
	newFile := hclwrite.NewEmptyFile()
	var resBlock *hclwrite.Block
	for _, b := range hclCfg.Body().Blocks() {
		if b.Type() == "resource" {
			resBlock = b
			continue
		}
		newFile.Body().AppendBlock(b)
	}
	if resBlock == nil {
		return "", fmt.Errorf("no resource block found")
	}
	resBody := resBlock.Body()
	resBody.SetAttributeValue("type", cty.StringVal(a.ResourceType()))
	for _, b := range resBody.Blocks() {
		if b.Type() == "dynamic" && b.Labels()[0] == "body" {
			resBody.RemoveBlock(b)
		}
	}
	variablePrefix := cfg.GetVariablePrefix(resourceTypeWithoutVendor(a.ResourceBlockType()))
	bodyVarName := "body"
	if variablePrefix != "" {
		bodyVarName = fmt.Sprintf("%s_body", variablePrefix)
	}
	bodyValue := newTokens().ident("var", 0).dot().ident(bodyVarName, 0).Tokens
	if cfg.Mode == UniVariable {
		uniVarName := variablePrefix
		if uniVarName == "" {
			uniVarName = resourceTypeWithoutVendor(a.ResourceBlockType())
		}
		bodyValue = newTokens().ident("var", 0).dot().ident(uniVarName, 0).dot().ident("body", 0).Tokens
	}

	resBody.SetAttributeRaw("body", bodyValue)
	newFile.Body().AppendBlock(resBlock)
	return string(newFile.Bytes()), nil
}

func (a azApiResourceGenerateCommand) ResourceType() string {
	return fmt.Sprintf("%s@%s", a.resourceType, a.apiVersion)
}

func (a azApiResourceGenerateCommand) ResourceBlockType() string {
	return "azapi_resource"
}

func (a azApiResourceGenerateCommand) Config() Config {
	return a.cfg
}

func (a azApiResourceGenerateCommand) Schema() (*tfjson.Schema, error) {
	resourceDef, err := azapi.GetAzApiType(a.resourceType, a.apiVersion)
	if err != nil {
		return nil, err
	}
	if resourceDef == nil {
		return nil, fmt.Errorf("unable to find resource definition for %s@%s", a.resourceType, a.apiVersion)
	}
	bodyType, ok := resourceDef.Body.Type.(*types.ObjectType)
	if !ok {
		return nil, fmt.Errorf("resource body type is not an object type")
	}
	blockSchema, err := azapi.ConvertAzApiObjectTypeToTerraformJsonSchemaAttribute(types.ObjectProperty{
		Type: &types.TypeReference{
			Type: bodyType,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert az api object type to terraform json schema: %+v", err)
	}
	a.enrichFields(blockSchema.Attributes)

	return &tfjson.Schema{
		Version: 0,
		Block:   blockSchema,
	}, nil
}

func (a azApiResourceGenerateCommand) enrichFields(fields map[string]*tfjson.SchemaAttribute) {
	fields["parent_id"] = &tfjson.SchemaAttribute{
		AttributeType:   cty.String,
		Description:     "The ID of the azure resource in which this resource is created.",
		DescriptionKind: tfjson.SchemaDescriptionKindPlain,
		Required:        true,
	}
}
