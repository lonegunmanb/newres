package pkg

import (
	"fmt"
	"github.com/hashicorp/terraform-json"
)

var _ ResourceGenerateCommand = generalResource{}
var _ withDocument = generalResource{}

type generalResource struct {
	resourceType string
	cfg          Config
}

func (g generalResource) ResourceType() string {
	return g.resourceType
}

func (g generalResource) Doc() (map[string]argumentDescription, error) {
	return newDocument(g.resourceType).parseDocument()
}

func (g generalResource) ResourceBlockType() string {
	return g.resourceType
}

func (g generalResource) Config() Config {
	return g.cfg
}

func (g generalResource) Schema() (*tfjson.Schema, error) {
	s, ok := resourceSchemas[g.resourceType]
	if !ok {
		return nil, fmt.Errorf("unsupported type %s", g.resourceType)
	}
	return s, nil
}
