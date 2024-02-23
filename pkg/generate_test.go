package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	azurermschema_v2 "github.com/lonegunmanb/terraform-azurerm-schema/v2/generated"
	azurermschema "github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestGenerateResourceBlock_InvalidResourcTypeShouldReturnError(t *testing.T) {
	_, err := GenerateResource("invalidType", Config{})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "unsupported type")
}

func TestGenerateResource_SimpleUniVarResource(t *testing.T) {
	resourceType := "azurerm_resource_group"
	schema := azurermschema.Resources[resourceType]
	generated, err := GenerateResource(resourceType, Config{
		Mode: UniVariable,
	})
	require.NoError(t, err)
	config, diag := hclsyntax.ParseConfig([]byte(generated), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	mod := tfconfig.NewModule("")
	diag = tfconfig.LoadModuleFromFile(config, mod)
	require.False(t, diag.HasErrors())
	assert.Equal(t, 2, len(config.Body.(*hclsyntax.Body).Blocks))
	assert.Equal(t, 1, len(mod.Variables))
	assert.Contains(t, mod.Variables, "resource_group")
	resourceAddress := fmt.Sprintf("%s.this", resourceType)
	assert.Contains(t, mod.ManagedResources, resourceAddress)
	for name := range schema.Block.Attributes {
		if name == "id" {
			continue
		}
		assert.Contains(t, generated, fmt.Sprintf("- `%s` -", name))
		assert.Contains(t, generated, fmt.Sprintf("%s = var.resource_group.%s", name, name))
	}
	for name := range schema.Block.NestedBlocks {
		assert.Contains(t, generated, fmt.Sprintf("`%s` block supports the following:", name))
		assert.Contains(t, generated, fmt.Sprintf(`dynamic "%s" {`, name))
	}
}

func TestGenerateResource_ObjectInAttributeShouldGenerateNestedBlock(t *testing.T) {
	code, err := GenerateResource("azurerm_container_group", Config{
		Mode: MultipleVariables,
	})
	require.NoError(t, err)
	assert.Contains(t, code, `dynamic "exposed_port" {`)
}

func TestGenerateResource_NestedObjectAsAttribute(t *testing.T) {
	// https://github.com/lonegunmanb/terraform-azurerm-schema/blob/main/generated/resource/azurermSiteRecoveryReplicatedVm.go#L33-L42
	resource := azurermschema.Resources["azurerm_site_recovery_replicated_vm"]
	attr := resource.Block.Attributes["managed_disk"]
	sut := restoreToNestedBlockSchema(attr)
	assert.Contains(t, sut.Block.NestedBlocks, "target_disk_encryption")
	assert.Contains(t, sut.Block.NestedBlocks["target_disk_encryption"].Block.NestedBlocks, "disk_encryption_key")
}

func TestGenerateResource_IdAttributeInsideNestedBlockAttributeShouldNotBeSkipped(t *testing.T) {
	resource := azurermschema.Resources["azurerm_storage_table"]
	nb := resource.Block.NestedBlocks["acl"]
	res, err := newResourceBlock("azurerm_storage_table", resource, Config{})
	require.NoError(t, err)
	n := newNestedBlock(res, "acl", nb)
	assert.Equal(t, 1, len(n.attrs))
	assert.Equal(t, "id", n.attrs[0].name)
}

func TestGenerateVariableTypeForWholeResource(t *testing.T) {
	// we're using v2 resource since it's stable now and won't be changed
	schema := azurermschema_v2.Resources["azurerm_site_recovery_replicated_vm"]
	r, err := newResourceBlock("azurerm_site_recovery_replicated_vm", schema, Config{})
	require.NoError(t, err)
	variableType := generateVariableType(r, true)
	//`managed_disk` and `network_interface` are `SchemaConfigModeAttr` so schema info was lost, we cannot know whether their attributes are optional or not. https://github.com/hashicorp/terraform-provider-azurerm/blob/v2.99.0/internal/services/recoveryservices/site_recovery_replicated_vm_resource.go#L118-L120
	expected := `object({
  name = string
  recovery_replication_policy_id = string
  recovery_vault_name = string
  resource_group_name = string
  source_recovery_fabric_name = string
  source_recovery_protection_container_name = string
  source_vm_id = string
  target_availability_set_id = optional(string)
  target_network_id = optional(string)
  target_recovery_fabric_id = string
  target_recovery_protection_container_id = string
  target_resource_group_id = string
  managed_disk = optional(set(object({
    disk_id = string
    staging_storage_account_id = string
    target_disk_encryption_set_id = string
    target_disk_type = string
    target_replica_disk_type = string
    target_resource_group_id = string
  })))
  network_interface = optional(set(object({
    recovery_public_ip_address_id = string
    source_network_interface_id = string
    target_static_ip = string
    target_subnet_name = string
  })))
  timeouts = optional(object({
    create = optional(string)
    delete = optional(string)
    read = optional(string)
    update = optional(string)
  }))
})`
	assert.Equal(t, strings.ReplaceAll(expected, " ", ""), strings.ReplaceAll(variableType, " ", ""))
}

func TestGenerateResource_SkippedAttributeShouldNotAppearInVariableDescription(t *testing.T) {
	cases := []struct {
		resourceType string
		caseName     string
	}{
		{
			resourceType: "azurerm_resource_group",
			caseName:     "id",
		},
		{
			resourceType: "azurerm_kubernetes_cluster",
			caseName:     "fqdn",
		},
	}
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		t.Run(fmt.Sprintf("%s.%s", c.resourceType, c.caseName), func(t *testing.T) {
			resourceType := c.resourceType
			generated, err := GenerateResource(resourceType, Config{
				Mode: UniVariable,
			})
			require.NoError(t, err)
			assert.NotContains(t, generated, fmt.Sprintf("- `%s` -", c.caseName))
		})
	}
}
