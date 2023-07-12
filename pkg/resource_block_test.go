package pkg

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/lonegunmanb/azure-verified-module-fix/pkg"
	awsschema "github.com/lonegunmanb/terraform-aws-schema/v5/generated"
	azurermschema "github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateVariableBlock_AzureRootRequiredArgument(t *testing.T) {
	expected := `
variable "resource_group_name" {
  type        = string
  nullable    = false
}
`
	expectedBlock := toVariableBlock(expected)
	s := azurermschema.Resources["azurerm_resource_group"]
	name := s.Block.Attributes["name"]
	dummyDefaultValues := map[string]argumentDescription{}
	r, err := newResourceBlock("azurerm_resource_group", s)
	require.NoError(t, err)
	block := r.schemaAttributeToHCLBlock("name", name, dummyDefaultValues)
	assert.NoError(t, err)
	actualBlock := toVariableBlock(string(block.BuildTokens(hclwrite.Tokens{}).Bytes()))
	expectedString := variableBlockToHclCode(expectedBlock)
	actualString := variableBlockToHclCode(actualBlock)
	assert.Equal(t, expectedString, actualString)
}

func TestGenerateVariableBlock_AwsRootRequiredArgument(t *testing.T) {
	expected := `
variable "vpc_cidr_block" {
  type        = string
  default     = null
}
`
	expectedBlock := toVariableBlock(expected)
	s := awsschema.Resources["aws_vpc"]
	name := s.Block.Attributes["cidr_block"]
	dummyDefaultValues := map[string]argumentDescription{}
	r, err := newResourceBlock("aws_vpc", s)
	block := r.schemaAttributeToHCLBlock("cidr_block", name, dummyDefaultValues)
	assert.NoError(t, err)
	actualBlock := toVariableBlock(string(block.BuildTokens(hclwrite.Tokens{}).Bytes()))
	expectedString := variableBlockToHclCode(expectedBlock)
	actualString := variableBlockToHclCode(actualBlock)
	assert.Equal(t, expectedString, actualString)
}

func TestGenerateVariableType_ComplexObject(t *testing.T) {
	containerAppSchema := azurermschema.Resources["azurerm_container_app"]
	input := containerAppSchema.Block.NestedBlocks["template"]
	r, err := newResourceBlock("azurerm_container_app", containerAppSchema)
	require.NoError(t, err)
	actual := strings.Replace(generateVariableType(newNestedBlock(r, "template", input), true), " ", "", -1)
	expected := strings.Replace(strings.Replace(`object({
  max_replicas = optional(number)
  min_replicas = optional(number)
  revision_suffix = optional(string)
  container = list(object({
    args = optional(list(string))
    command = optional(list(string))
    cpu = number
    image = string
    memory = string
    name = string
    env = optional(list(object({
      name = string
      secret_name = optional(string)
      value = optional(string)
    })))
    liveness_probe = optional(list(object({
	  failure_count_threshold = optional(number)
	  host = optional(string)
	  initial_delay = optional(number)
      interval_seconds = optional(number)
	  path = optional(string)
 	  port = number
	  timeout = optional(number)
	  transport = string
      header = optional(list(object({
		name = string
		value = string
	  })))
    })))
    readiness_probe = optional(list(object({
	  failure_count_threshold = optional(number)
	  host = optional(string)
	  interval_seconds = optional(number)
	  path = optional(string)
	  port = number
	  success_count_threshold = optional(number)
	  timeout = optional(number)
	  transport = string
	  header = optional(list(object({
		name = string
		value = string
	  })))
	})))
	startup_probe = optional(list(object({
	  failure_count_threshold = optional(number)
	  host = optional(string)
	  interval_seconds = optional(number)
	  path = optional(string)
	  port = number
	  timeout = optional(number)
	  transport = string
	  header = optional(list(object({
		name = string
		value = string
	  })))
	})))
	volume_mounts = optional(list(object({
	  name = string
	  path = string
	})))
  }))
  volume = optional(list(object({
	name = string
	storage_name = optional(string)
	storage_type = optional(string)
  })))
})`, " ", "", -1), "	", "", -1)
	assert.Equal(t, expected, actual)
}

func TestGenerateVariableType_ListOfSimpleObject(t *testing.T) {
	containerAppSchema := azurermschema.Resources["azurerm_container_app"]
	input := containerAppSchema.Block.NestedBlocks["template"].
		Block.NestedBlocks["container"].
		Block.NestedBlocks["env"]
	r, err := newResourceBlock("azurerm_container_app", containerAppSchema)
	require.NoError(t, err)
	actual := strings.Replace(generateVariableType(newNestedBlock(r, "template", input), true), " ", "", -1)
	expected := strings.Replace(strings.Replace(`list(object({
  name = string
  secret_name = optional(string)
  value = optional(string)
}))`, " ", "", -1), "	", "", -1)
	assert.Equal(t, expected, actual)
}

func TestGenerateVariableBlock_RootArgumentDescription(t *testing.T) {
	resourceType := "azurerm_kubernetes_cluster"
	r, _ := newResourceBlock(resourceType, resourceSchemas[resourceType])
	desc := "(Required) The name of the Managed Kubernetes Cluster to create. Changing this forces a new resource to be created."
	generated, err := r.generateResource(map[string]argumentDescription{
		"name": {
			name:         "name",
			desc:         desc,
			defaultValue: nil,
		},
	}, true, multiVarsAttributeExpr, multiVarsNestedBlockIterator)
	require.NoError(t, err)
	config, diag := hclsyntax.ParseConfig([]byte(generated), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	mod := tfconfig.NewModule("")
	diag = tfconfig.LoadModuleFromFile(config, mod)
	require.False(t, diag.HasErrors())
	assert.Equal(t, desc, mod.Variables["kubernetes_cluster_name"].Description)
}

func variableBlockToHclCode(b *pkg.VariableBlock) string {
	f := hclwrite.NewFile()
	f.Body().AppendBlock(b.Block.WriteBlock)
	return string(f.Bytes())
}

func toVariableBlock(variableCode string) *pkg.VariableBlock {
	f, _ := pkg.ParseConfig([]byte(variableCode), "")
	b := f.GetBlock(0)
	variableBlock := pkg.BuildVariableBlock(f.File, pkg.NewHclBlock(b.Block, b.WriteBlock))
	variableBlock.AutoFix()
	return variableBlock
}
