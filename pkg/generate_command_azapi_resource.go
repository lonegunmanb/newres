package pkg

import (
	"fmt"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	"github.com/ms-henglu/go-azure-types/types"
)

var _ ResourceGenerateCommand = azApiResourceGenerateCommand{}

type azApiResourceGenerateCommand struct {
	ResourceType string
	ApiVersion   string
	Cfg          Config
}

func (a azApiResourceGenerateCommand) Type() string {
	return fmt.Sprintf("azapi-%s@%s", a.ResourceType, a.ApiVersion)
}

func (a azApiResourceGenerateCommand) Config() Config {
	return a.Cfg
}

func (a azApiResourceGenerateCommand) Schema() (*tfjson.Schema, error) {
	resourceDef, err := azapi.GetAzApiType(a.ResourceType, a.ApiVersion)
	if err != nil {
		return nil, err
	}
	if resourceDef == nil {
		return nil, fmt.Errorf("unable to find resource definition for %s@%s", a.ResourceType, a.ApiVersion)
	}
	bodyType, ok := resourceDef.Body.Type.(*types.ObjectType)
	if !ok {
		return nil, fmt.Errorf("resource body type is not an object type")
	}
	schemaAttribute := azapi.ConvertAzApiTypeToTerraformJsonSchemaAttribute(types.ObjectProperty{
		Type: &types.TypeReference{
			Type: bodyType,
		},
	})
	panic(schemaAttribute)
}
