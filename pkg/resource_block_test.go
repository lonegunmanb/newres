package pkg

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	avmfix "github.com/lonegunmanb/avmfix/pkg"
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
	r, err := newResourceBlock("azurerm_resource_group", s, Config{})
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
	r, err := newResourceBlock("aws_vpc", s, Config{})
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
	r, err := newResourceBlock("azurerm_container_app", containerAppSchema, Config{})
	require.NoError(t, err)
	actual := strings.Replace(generateVariableType(newNestedBlock(r, "template", input), true), " ", "", -1)
	expected := strings.Replace(strings.Replace(`object({
    max_replicas           = optional(number)
    min_replicas           = optional(number)
    revision_suffix        = optional(string)
    azure_queue_scale_rule = optional(list(object({
      name           = string
      queue_length   = number
      queue_name     = string
      authentication = list(object({
        secret_name       = string
        trigger_parameter = string
      }))
    })))
    container = list(object({
      args    = optional(list(string))
      command = optional(list(string))
      cpu     = number
      image   = string
      memory  = string
      name    = string
      env     = optional(list(object({
        name        = string
        secret_name = optional(string)
        value       = optional(string)
      })))
      liveness_probe = optional(list(object({
        failure_count_threshold = optional(number)
        host                    = optional(string)
        initial_delay           = optional(number)
        interval_seconds        = optional(number)
        path                    = optional(string)
        port                    = number
        timeout                 = optional(number)
        transport = string
        header    = optional(list(object({
          name  = string
          value = string
        })))
      })))
      readiness_probe = optional(list(object({
        failure_count_threshold = optional(number)
        host                    = optional(string)
        interval_seconds        = optional(number)
        path                    = optional(string)
        port                    = number
        success_count_threshold = optional(number)
        timeout                 = optional(number)
        transport               = string
        header                  = optional(list(object({
          name  = string
          value = string
        })))
      })))
      startup_probe = optional(list(object({
        failure_count_threshold = optional(number)
        host                    = optional(string)
        interval_seconds        = optional(number)
        path                    = optional(string)
        port                    = number
        timeout                 = optional(number)
        transport               = string
        header                  = optional(list(object({
          name  = string
          value = string
        })))
      })))
      volume_mounts = optional(list(object({
        name = string
        path = string
      })))
    }))
    custom_scale_rule = optional(list(object({
      custom_rule_type = string
      metadata         = map(string)
      name             = string
      authentication   = optional(list(object({
        secret_name       = string
        trigger_parameter = string
      })))
    })))
    http_scale_rule = optional(list(object({
      concurrent_requests = string
      name                = string
      authentication      = optional(list(object({
        secret_name       = string
        trigger_parameter = optional(string)
      })))
    })))
    init_container = optional(list(object({
      args    = optional(list(string))
      command = optional(list(string))
      cpu     = optional(number)
      image   = string
      memory  = optional(string)
      name    = string
      env     = optional(list(object({
        name        = string
        secret_name = optional(string)
        value       = optional(string)
      })))
      volume_mounts = optional(list(object({
        name = string
        path = string
      })))
    })))
    tcp_scale_rule = optional(list(object({
      concurrent_requests = string
      name                = string
      authentication      = optional(list(object({
        secret_name       = string
        trigger_parameter = optional(string)
      })))
    })))
    volume = optional(list(object({
      name         = string
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
	r, err := newResourceBlock("azurerm_container_app", containerAppSchema, Config{})
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
	r, _ := newResourceBlock(resourceType, resourceSchemas[resourceType], Config{})
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

func TestGenerateVariableBlock_CustomizedHeredocDelimiter(t *testing.T) {
	resourceType := "azurerm_kubernetes_cluster"
	r, _ := newResourceBlock(resourceType, resourceSchemas[resourceType], Config{
		Delimiter: "DOCUMENT",
	})
	desc := "(Required) The name of the Managed Kubernetes Cluster to create. Changing this forces a new resource to be created."
	generated, err := r.generateUniVarResource(map[string]argumentDescription{
		"name": {
			name:         "name",
			desc:         desc,
			defaultValue: nil,
		},
	})
	require.NoError(t, err)
	assert.Contains(t, generated, "description = <<-DOCUMENT")
}

func TestGenerateResourceBlock_ComputedOnlyAttributeShouldNotInGeneratedResourceBlock(t *testing.T) {
	resourceType := "azurerm_web_app_hybrid_connection"
	r, _ := newResourceBlock(resourceType, resourceSchemas[resourceType], Config{
		Delimiter: "DOCUMENT",
	})
	cases := []struct {
		attributeName string
		want          bool
	}{
		{
			attributeName: "namespace_name",
			want:          false,
		},
		{
			// Required
			attributeName: "port",
			want:          true,
		},
		{
			// Optional
			attributeName: "send_key_name",
			want:          true,
		},
	}
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		t.Run(c.attributeName, func(t *testing.T) {
			for _, a := range r.attrs {
				if a.name == c.attributeName {
					if c.want {
						return
					} else {
						t.Fatal("computed only attribute should be excluded.")
					}
				}
			}
			if c.want {
				t.Fatal("no computed only attribute should not be excluded.")
			}
		})
	}

}

func variableBlockToHclCode(b *avmfix.VariableBlock) string {
	f := hclwrite.NewFile()
	f.Body().AppendBlock(b.Block.WriteBlock)
	return string(f.Bytes())
}

func toVariableBlock(variableCode string) *avmfix.VariableBlock {
	f, _ := avmfix.ParseConfig([]byte(variableCode), "")
	b := f.GetBlock(0)
	variableBlock := avmfix.BuildVariableBlock(f.File, avmfix.NewHclBlock(b.Block, b.WriteBlock))
	variableBlock.AutoFix()
	return variableBlock
}
