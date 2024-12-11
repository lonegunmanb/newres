package pkg

import tfjson "github.com/hashicorp/terraform-json"

type ResourceGenerateCommand interface {
	Type() string
	Config() Config
	Schema() (*tfjson.Schema, error)
}

func NewResourceGenerateCommand(resourceType string, cfg Config, parameters map[string]string) ResourceGenerateCommand {
	return generalResource{
		ResourceType: resourceType,
		Cfg:          cfg,
	}
}
