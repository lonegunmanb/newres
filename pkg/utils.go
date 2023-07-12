package pkg

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ahmetb/go-linq/v3"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

func resourceTypeValid(resourceType string) bool {
	s := strings.Split(resourceType, "_")
	return len(s) > 1
}

func resourceVendor(resourceType string) string {
	return strings.Split(resourceType, "_")[0]
}

func resourceTypeWithoutVendor(resourceType string) string {
	return strings.TrimPrefix(resourceType, fmt.Sprintf("%s_", resourceVendor(resourceType)))
}

func inferNestingMode(attributeType cty.Type) tfjson.SchemaNestingMode {
	nestingMode := tfjson.SchemaNestingModeSingle
	if attributeType.IsSetType() {
		nestingMode = tfjson.SchemaNestingModeSet
	} else if attributeType.IsListType() {
		nestingMode = tfjson.SchemaNestingModeList
	} else if attributeType.IsMapType() {
		nestingMode = tfjson.SchemaNestingModeMap
	}
	return nestingMode
}

func ctyTypeToVariableTypeString(t cty.Type) string {
	switch t {
	case cty.String:
		return "string"
	case cty.Number:
		return "number"
	case cty.Bool:
		return "bool"
	}
	if t.SetElementType() != nil {
		return fmt.Sprintf("set(%s)", ctyTypeToVariableTypeString(t.ElementType()))
	}
	if t.ListElementType() != nil {
		return fmt.Sprintf("list(%s)", ctyTypeToVariableTypeString(t.ElementType()))
	}
	if t.MapElementType() != nil {
		return fmt.Sprintf("map(%s)", ctyTypeToVariableTypeString(t.ElementType()))
	}
	if t.IsObjectType() {
		sb := strings.Builder{}
		var attributes []struct {
			name string
			t    cty.Type
		}
		linq.From(t.AttributeTypes()).OrderBy(func(i interface{}) interface{} {
			return i.(linq.KeyValue).Key
		}).Select(func(i any) any {
			pair := i.(linq.KeyValue)
			return struct {
				name string
				t    cty.Type
			}{name: pair.Key.(string), t: pair.Value.(cty.Type)}
		}).ToSlice(&attributes)

		for _, pair := range attributes {
			fieldType := ctyTypeToVariableTypeString(pair.t)
			sb.WriteString(fmt.Sprintf("%s = %s\n", pair.name, fieldType))
		}
		return fmt.Sprintf(`object({
  %s})`, sb.String())
	}
	panic(fmt.Sprintf("unexpected type: %s", t.FriendlyName()))
}

func fetchURLContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch content from %s, status code: %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
