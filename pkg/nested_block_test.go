package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	azurermschema "github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNestedBlock_SimpleObject(t *testing.T) {
	resourceSchema := azurermschema.Resources["azurerm_kubernetes_cluster"]
	identity := resourceSchema.Block.NestedBlocks["identity"]
	r, err := newResourceBlock("azurerm_shared_image", resourceSchema)
	require.NoError(t, err)
	sut := newNestedBlock(r, "identity", identity)
	actual := sut.generateHCL("var.kubernetes_cluster")
	expected := `
dynamic "identity" {
  for_each = var.kubernetes_cluster.identity == null ? [] : [var.kubernetes_cluster.identity]
  content {
    identity_ids = identity.value.identity_ids
    type = identity.value.type
  }
}
`
	expected = formatHcl(expected, t)
	actual = formatHcl(actual, t)
	assert.Equal(t, expected, actual)
}

func TestGenerateNestedBlock_SimpleList(t *testing.T) {
	schema := azurermschema.Resources["azurerm_container_app"]
	env := schema.
		Block.NestedBlocks["template"].
		Block.NestedBlocks["container"].
		Block.NestedBlocks["env"]
	r, err := newResourceBlock("azurerm_container_app", schema)
	require.NoError(t, err)
	sut := newNestedBlock(r, "env", env)
	actual := sut.generateHCL("container.value")
	expected := `
dynamic "env" {
  for_each = container.value.env == null ? [] : container.value.env
  content {
    name = env.value.name
	secret_name = env.value.secret_name
	value = env.value.value
  }
}
`
	expected = formatHcl(expected, t)
	actual = formatHcl(actual, t)
	assert.Equal(t, expected, actual)
}

func TestGenerateNestedBlock_ListOfObject(t *testing.T) {
	schema := azurermschema.Resources["azurerm_container_app"]
	container := schema.
		Block.NestedBlocks["template"].
		Block.NestedBlocks["container"]
	r, err := newResourceBlock("azurerm_container_app", schema)
	require.NoError(t, err)
	sut := newNestedBlock(r, "container", container)
	actual := sut.generateHCL("template.value")
	assert.Contains(t, actual, "for_each = template.value.container")
}

func TestGenerateNestedBlock_SimpleSet(t *testing.T) {
	schema := azurermschema.Resources["azurerm_container_app"]
	secret := schema.
		Block.NestedBlocks["secret"]
	r, err := newResourceBlock("azurerm_container_app", schema)
	require.NoError(t, err)
	sut := newNestedBlock(r, "secret", secret)
	actual := sut.generateHCL("var.container_app")
	expected := `
dynamic "secret" {
  for_each = var.container_app.secret == null ? [] : var.container_app.secret
  content {
    name = secret.value.name
	value = secret.value.value
  }
}
`
	expected = formatHcl(expected, t)
	actual = formatHcl(actual, t)
	assert.Equal(t, expected, actual)
}

func TestGenerateNestedBlock_ContainerGroupContainer(t *testing.T) {
	schema := azurermschema.Resources["azurerm_container_group"]
	container := schema.
		Block.NestedBlocks["container"]
	r, err := newResourceBlock("azurerm_container_group", schema)
	require.NoError(t, err)
	sut := newNestedBlock(r, "container", container)
	actual := sut.generateHCL("var.container_group")
	expected := `
dynamic "container" {
  for_each = var.container_group.container
  content {
	commands = container.value.commands
	cpu = container.value.cpu
	cpu_limit = container.value.cpu_limit
	environment_variables = container.value.environment_variables
	image = container.value.image
	memory = container.value.memory
	memory_limit = container.value.memory_limit
    name = container.value.name
	secure_environment_variables = container.value.secure_environment_variables
	dynamic "gpu" {
	  for_each = container.value.gpu == null ? [] : [container.value.gpu]
	  content {
		count = gpu.value.count
		sku = gpu.value.sku
	  }
	}
	dynamic "gpu_limit" {
	  for_each = container.value.gpu_limit == null ? [] : [container.value.gpu_limit]
	  content {
		count = gpu_limit.value.count
		sku = gpu_limit.value.sku
	  }
	}
	dynamic "liveness_probe" {
	  for_each = container.value.liveness_probe == null ? [] : [container.value.liveness_probe]
	  content {
		exec = liveness_probe.value.exec
		failure_threshold = liveness_probe.value.failure_threshold
		initial_delay_seconds = liveness_probe.value.initial_delay_seconds
		period_seconds = liveness_probe.value.period_seconds
		success_threshold = liveness_probe.value.success_threshold
		timeout_seconds = liveness_probe.value.timeout_seconds
		dynamic "http_get" {
		  for_each = liveness_probe.value.http_get == null ? [] : liveness_probe.value.http_get
		  content {
			http_headers = http_get.value.http_headers
			path = http_get.value.path
			port = http_get.value.port
			scheme = http_get.value.scheme
		  }
		}
	  }
	}
	dynamic "ports" {
	  for_each = container.value.ports == null ? [] : container.value.ports
	  content {
		port = ports.value.port
		protocol = ports.value.protocol
	  }
	}
	dynamic "readiness_probe" {
	  for_each = container.value.readiness_probe == null ? [] : [container.value.readiness_probe]
	  content {
		exec = readiness_probe.value.exec
		failure_threshold = readiness_probe.value.failure_threshold
		initial_delay_seconds = readiness_probe.value.initial_delay_seconds
		period_seconds = readiness_probe.value.period_seconds
		success_threshold = readiness_probe.value.success_threshold
		timeout_seconds = readiness_probe.value.timeout_seconds
		dynamic "http_get" {
		  for_each = readiness_probe.value.http_get == null ? [] : readiness_probe.value.http_get
		  content {
			http_headers = http_get.value.http_headers
			path = http_get.value.path
			port = http_get.value.port
			scheme = http_get.value.scheme
		  }
		}
	  }
	}
	dynamic "volume" {
	  for_each = container.value.volume == null ? [] : container.value.volume
	  content {
		empty_dir = volume.value.empty_dir
		mount_path = volume.value.mount_path
		name = volume.value.name
		read_only = volume.value.read_only
		secret = volume.value.secret
		share_name = volume.value.share_name
		storage_account_key = volume.value.storage_account_key
		storage_account_name = volume.value.storage_account_name
		dynamic "git_repo" {
		  for_each = volume.value.git_repo == null ? [] : [volume.value.git_repo]
		  content {
			directory = git_repo.value.directory
			revision = git_repo.value.revision
			url = git_repo.value.url
		  }
		}
	  }
	}
  }
}
`
	expected = formatHcl(expected, t)
	actual = formatHcl(actual, t)
	assert.Equal(t, expected, actual)
}

func TestGenerateNestedBlock_RequiredObject(t *testing.T) {
	res := azurermschema.Resources["azurerm_shared_image"]
	identifier := res.Block.NestedBlocks["identifier"]
	r, err := newResourceBlock("azurerm_shared_image", res)
	require.NoError(t, err)
	sut := newNestedBlock(r, "identifier", identifier)
	actual := sut.generateHCL("var.shared_image_identifier")
	expected := `
dynamic "identifier" {
  for_each = [var.shared_image_identifier]
  content {
    offer = identifier.value.offer
    publisher = identifier.value.publisher
    sku = identifier.value.sku
  }
}
`
	expected = formatHcl(expected, t)
	actual = formatHcl(actual, t)
	assert.Equal(t, expected, actual)
}

func formatHcl(code string, t *testing.T) string {
	config, diag := hclwrite.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	f := hclwrite.NewFile()
	f.Body().AppendBlock(config.Body().Blocks()[0])
	return string(f.Bytes())
}

func TestGenerateVariableBlock_NestedBlockDescription(t *testing.T) {
	resourceType := "azurerm_kubernetes_cluster"
	r, _ := newResourceBlock(resourceType, resourceSchemas[resourceType])
	desc := "(Required) The subnet name for the virtual nodes to run."
	generated, err := r.generateResource(map[string]argumentDescription{
		"aci_connector_linux.subnet_name": {
			name: "subnet_name",
			desc: desc,
		},
	}, true, multiVarsAttributeExpr, multiVarsNestedBlockIterator)
	require.NoError(t, err)
	config, diag := hclsyntax.ParseConfig([]byte(generated), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	mod := tfconfig.NewModule("")
	diag = tfconfig.LoadModuleFromFile(config, mod)
	require.False(t, diag.HasErrors())
	actual := mod.Variables["kubernetes_cluster_aci_connector_linux"].Description
	expected := fmt.Sprintf("- `subnet_name` - %s", desc)
	actual = strings.TrimSuffix(actual, "\n")
	expected = strings.TrimPrefix(expected, "\n")
	assert.Equal(t, expected, actual)
}

func TestGenerateVariableType_SimpleObject(t *testing.T) {
	aksSchema := azurermschema.Resources["azurerm_kubernetes_cluster"]
	r, err := newResourceBlock("azurerm_kubernetes_cluster", aksSchema)
	cases := []struct {
		nestedBlockName string
		expected        string
	}{
		{
			nestedBlockName: "confidential_computing",
			expected: `object({
  sgx_quote_helper_enabled = bool
})`,
		},
		{
			nestedBlockName: "identity",
			expected: `object({
  identity_ids = optional(set(string))
  type = string
})`,
		},
	}
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		t.Run(c.nestedBlockName, func(t *testing.T) {
			input := aksSchema.Block.NestedBlocks[c.nestedBlockName]
			require.NoError(t, err)
			actual := strings.Replace(generateVariableType(newNestedBlock(r, c.nestedBlockName, input), true), " ", "", -1)
			assert.Equal(t, strings.Replace(c.expected, " ", "", -1), actual)
		})
	}
}

func TestGenerateVariableType_RequiredObject(t *testing.T) {
	input := azurermschema.Resources["azurerm_kubernetes_cluster"]
	resourceBlock, _ := newResourceBlock("azurerm_kubernetes_cluster", input)
	actual := strings.Replace(generateVariableType(resourceBlock, true), " ", "", -1)
	assert.Contains(t, actual, "default_node_pool=object({")
}

func TestGenerateDynamicBlockForAzurermTimeouts(t *testing.T) {
	code, err := GenerateResource("azurerm_storage_table", MultipleVariables)
	require.NoError(t, err)
	assert.Contains(t, code, "for_each = var.storage_table_timeouts == null ? [] : [var.storage_table_timeouts]")
}
