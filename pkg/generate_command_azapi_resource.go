package pkg

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

var _ ResourceGenerateCommand = azApiResourceGenerateCommand{}

type azApiResourceGenerateCommand struct {
	resourceType string
	apiVersion   string
	cfg          Config
}

func (a azApiResourceGenerateCommand) ResourceType() string {
	return fmt.Sprintf("azapi-%s@%s", a.resourceType, a.apiVersion)
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
