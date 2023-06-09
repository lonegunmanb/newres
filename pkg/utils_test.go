package pkg

import (
	azurermschema "github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGenerateVariableType_ObjectTypeInAttributes(t *testing.T) {
	containerAppSchema := azurermschema.Resources["azurerm_container_group"]
	input := containerAppSchema.Block.Attributes["exposed_port"]
	actual := strings.Replace(ctyTypeToVariableTypeString(input.AttributeType), " ", "", -1)
	expected := strings.Replace(`set(object({
  port = number
  protocol = string
}))`, " ", "", -1)
	assert.Equal(t, expected, actual)
}
