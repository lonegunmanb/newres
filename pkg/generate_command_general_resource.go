package pkg

import (
	"fmt"
	"github.com/hashicorp/terraform-json"
)

var _ ResourceGenerateCommand = generalResource{}
var _ withDocument = generalResource{}

type generalResource struct {
	ResourceType string
	Cfg          Config
}

func (g generalResource) Doc() (map[string]argumentDescription, error) {
	return newDocument(g.ResourceType).parseDocument()
}

func (g generalResource) Type() string {
	return g.ResourceType
}

func (g generalResource) Config() Config {
	return g.Cfg
}

func (g generalResource) Schema() (*tfjson.Schema, error) {
	s, ok := resourceSchemas[g.ResourceType]
	if !ok {
		return nil, fmt.Errorf("unsupported type %s", g.ResourceType)
	}
	return s, nil
}
