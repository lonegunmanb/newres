package pkg

import (
	"fmt"
)

func GenerateResource(generateCmd ResourceGenerateCommand) (string, error) {
	resourceType := generateCmd.Type()
	cfg := generateCmd.Config()
	schema, err := generateCmd.Schema()
	if err != nil {
		return "", err
	}
	r, err := newResourceBlock(resourceType, schema, cfg)
	if err != nil {
		return "", fmt.Errorf("error on parse resource type name %s: %s", resourceType, err.Error())
	}
	post, ok := generateCmd.(postProcessor)
	if ok {
		post.action(r)
	}
	document := make(map[string]argumentDescription)
	docGenerate, ok := generateCmd.(withDocument)
	if ok {
		document, err = docGenerate.Doc()
	}
	if err != nil {
		return "", fmt.Errorf("error on load and parse document: %s", err.Error())
	}
	if cfg.GetMode() == UniVariable {
		return r.generateUniVarResource(document)
	}
	return r.generateMultiVarsResource(document)
}
