package pkg

import (
	"fmt"
)

func GenerateResource(resourceType string, cfg Config) (string, error) {
	schema, ok := resourceSchemas[resourceType]
	if !ok {
		return "", fmt.Errorf("unsupported type %s", resourceType)
	}
	r, err := newResourceBlock(resourceType, schema, cfg)
	if err != nil {
		return "", fmt.Errorf("error on parse resource type name %s: %s", resourceType, err.Error())
	}
	document, err := newDocument(r.name).parseDocument()
	if err != nil {
		return "", fmt.Errorf("error on load and parse document: %s", err.Error())
	}
	if cfg.GetMode() == UniVariable {
		return r.generateUniVarResource(document)
	}
	return r.generateMultiVarsResource(document)
}
