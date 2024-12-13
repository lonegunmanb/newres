package pkg

import (
	"fmt"
)

func GenerateResource(generateCmd ResourceGenerateCommand) (string, error) {
	blockType := generateCmd.ResourceBlockType()
	cfg := generateCmd.Config()
	schema, err := generateCmd.Schema()
	if err != nil {
		return "", err
	}
	r, err := newResourceBlock(blockType, schema, cfg)
	if err != nil {
		return "", fmt.Errorf("error on parse resource type name %s: %s", generateCmd.ResourceType(), err.Error())
	}
	document := make(map[string]argumentDescription)
	docGenerate, ok := generateCmd.(withDocument)
	if ok {
		document, err = docGenerate.Doc()
	}
	if err != nil {
		return "", fmt.Errorf("error on load and parse document: %s", err.Error())
	}
	var generated string
	if cfg.GetMode() == UniVariable {
		generated, err = r.generateUniVarResource(document)
	} else {
		generated, err = r.generateMultiVarsResource(document)
	}
	if err != nil {
		return "", err
	}
	post, ok := generateCmd.(postProcessor)
	if ok {
		generated, err = post.action(generated, cfg)
	}
	return generated, err
}
