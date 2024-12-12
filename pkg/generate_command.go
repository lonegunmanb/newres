package pkg

import (
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

const AzApiResourceType = "azapi-resource-type"

type ResourceGenerateCommand interface {
	ResourceBlockType() string
	ResourceType() string
	Config() Config
	Schema() (*tfjson.Schema, error)
}

func NewResourceGenerateCommand(resourceType string, cfg Config, parameters map[string]string) ResourceGenerateCommand {
	var g ResourceGenerateCommand = generalResource{
		resourceType: resourceType,
		cfg:          cfg,
	}
	if resourceType == "azapi_resource" {
		if azapiType, ok := parameters[AzApiResourceType]; ok {
			typeString := strings.Split(azapiType, "@")
			g = azApiResourceGenerateCommand{
				resourceType: typeString[0],
				apiVersion:   typeString[1],
				cfg:          cfg,
			}
		}
	}
	return g
}
