package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	autofix "github.com/lonegunmanb/avmfix/pkg"
	"github.com/lonegunmanb/newres/v3/pkg"
)

var azapiVersionRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+(/[a-zA-Z0-9.-]+)+@[0-9]{4}-[0-9]{2}-[0-9]{2}(-preview)?$`)

func main() {
	// Parse command line flags
	dir := flag.String("dir", "", "Directory path to store generated files (required)")
	univar := flag.Bool("u", false, "Generate mode: UniVariable if set, MultipleVariables if not set")
	resourceType := flag.String("r", "", "Resource type to generate configuration for (required)")
	delimiter := flag.String("delimiter", "EOT", "Heredoc delimiter (optional)")
	azapiResourceType := flag.String(pkg.AzApiResourceType, "", "AZAPI resource type (optional)")
	flag.StringVar(resourceType, "resource-type", "", "")
	flag.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, "Usage: newres -dir [DIRECTORY] [-u] [-r RESOURCE_TYPE] [-delimiter DELIMITER]")
		_, _ = fmt.Fprintln(os.Stderr, "       newres -dir [DIRECTORY] [-u] [--resource-type RESOURCE_TYPE] [-delimiter DELIMITER]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *dir == "" || *resourceType == "" {
		flag.Usage()
		os.Exit(1)
	}
	parameters := map[string]string{}

	// Check if resourceType is azapi and azapiResourceType is provided
	if *azapiResourceType != "" {
		if *resourceType != "azapi_resource" {
			fmt.Println("Error: --azapi-resource-type must be provided when --resource-type is `azapi_resource`")
			os.Exit(1)
		}
		if !azapiVersionRegex.MatchString(*azapiResourceType) {
			fmt.Println("Error: Invalid azapi-resource-type format")
			os.Exit(1)
		}
		parameters[pkg.AzApiResourceType] = *azapiResourceType
	}

	// Set generate mode based on the -u flag
	var generateMode pkg.GenerateMode
	if *univar {
		generateMode = pkg.UniVariable
	} else {
		generateMode = pkg.MultipleVariables
	}

	if delimiter == nil {
		empty := ""
		delimiter = &empty
	}

	// Call GenerateResource function
	generatedCode, err := pkg.GenerateResource(pkg.NewResourceGenerateCommand(*resourceType, pkg.Config{
		Delimiter: *delimiter,
		Mode:      generateMode,
	}, parameters))
	if err != nil {
		fmt.Printf("Error generating resource: %s\n", err)
		os.Exit(1)
	}

	// Split generated code into variable and resource blocks
	variablesFile := hclwrite.NewEmptyFile()
	resourceFile := hclwrite.NewEmptyFile()

	generatedFile, diag := hclwrite.ParseConfig([]byte(generatedCode), "", hcl.InitialPos)
	if diag.HasErrors() {
		fmt.Printf("Error parsing generated code: %s\n", diag.Error())
		os.Exit(1)
	}

	for _, block := range generatedFile.Body().Blocks() {
		switch block.Type() {
		case "variable":
			variablesFile.Body().AppendBlock(block)
			variablesFile.Body().AppendNewline()
		case "resource":
			resourceFile.Body().AppendBlock(block)
			resourceFile.Body().AppendNewline()
		}
	}

	// Append content to variables.tf and main.tf
	variablesPath := filepath.Join(*dir, "variables.tf")
	mainPath := filepath.Join(*dir, "main.tf")

	err = appendToFile(variablesPath, variablesFile.Bytes(), 0600)
	if err != nil {
		fmt.Printf("Error writing variables.tf: %s\n", err)
		os.Exit(1)
	}

	err = appendToFile(mainPath, resourceFile.Bytes(), 0600)
	if err != nil {
		fmt.Printf("Error writing main.tf: %s\n", err)
		os.Exit(1)
	}

	err = autofix.DirectoryAutoFix(*dir)
	if err != nil {
		fmt.Printf("Error autofix: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated variables.tf and main.tf")
}

func appendToFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer func() {

		_ = f.Close()
	}()

	_, err = f.Write(data)
	return err
}
