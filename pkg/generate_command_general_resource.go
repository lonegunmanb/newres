package pkg

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
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
	schema, err := getResourceSchema(g.resourceType, g.cfg.ProviderNamespace, g.cfg.ProviderVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for %s: %w", g.resourceType, err)
	}
	return schema, nil
}
