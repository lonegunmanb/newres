package pkg

import (
	"bufio"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"regexp"
	"strings"
)

var urlTemplates = map[string]string{
	"azurerm": "https://raw.githubusercontent.com/hashicorp/terraform-provider-azurerm/main/website/docs/r/%s.html.markdown",
	"azuread": "https://raw.githubusercontent.com/hashicorp/terraform-provider-azuread/main/docs/resources/%s.md",
	"aws":     "https://raw.githubusercontent.com/hashicorp/terraform-provider-aws/main/website/docs/r/%s.html.markdown",
	"google":  "https://raw.githubusercontent.com/hashicorp/terraform-provider-google/main/website/docs/r/%s.html.markdown",
}

var backQuoteNameRegexp = regexp.MustCompile(`\x60.+\x60`)
var argumentsHeadlineRegex = regexp.MustCompile("## [A|a]rguments? [R|r]eference")
var timeoutsHeadlineRegex = regexp.MustCompile("## [T|t]imeouts?")
var attributesHeadlineRegex = regexp.MustCompile("## [A|a]ttributes? [R|r]eference")
var importHeadlineRegex = regexp.MustCompile("## [I|i]mports?")
var awsNestedBlockDescriptionKeyWords = linq.From([]string{
	"following",
	"argument",
	"configuration block:",
})
var awsTimeoutsUrl = "https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts"
var azureTimeoutsUrl = "https://www.terraform.io/docs/configuration/resources.html#timeouts"
var googleTimeoutsUrl = "https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/retries-and-customizable-timeouts"
var timeoutsUrlSuffix = "resources/syntax#operation-timeouts"

type Document struct {
	resourceType string
	getContent   func(string) (string, error)
}

func newDocument(resourceType string) Document {
	return Document{
		resourceType: resourceType,
		getContent:   content,
	}
}

func (d Document) content() (string, error) {
	return d.getContent(d.resourceType)
}

var content = func(resourceType string) (string, error) {
	if !resourceTypeValid(resourceType) {
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	vendor := resourceVendor(resourceType)
	tplt, ok := urlTemplates[vendor]
	if !ok {
		return "", nil
	}
	return fetchURLContent(fmt.Sprintf(tplt, resourceTypeWithoutVendor(resourceType)))
}

func (d Document) parseDocument() (map[string]argumentDescription, error) {
	r := make(map[string]argumentDescription, 0)
	markdown, err := d.content()
	if err != nil {
		return nil, fmt.Errorf("error on get document: %s", err.Error())
	}
	markdown = strings.Replace(markdown, "\r\n", "\n", -1)
	// aws document's lines need extra line break
	markdown = strings.Replace(markdown, "\n*", "\n\n*", -1)
	scanner := bufio.NewScanner(strings.NewReader(markdown))

	parsing := false
	lineBuilder := strings.Builder{}
	blockName := ""
	// Iterate through the lines in the input string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if d.beginParse(line) {
			parsing = true
			continue
		} else if d.endParse(line) {
			parsing = false
			continue
		}

		if parsing {
			if line != "" {
				if lineBuilder.Len() > 0 {
					lineBuilder.WriteString(" ")
				}
				lineBuilder.WriteString(line)
			} else {
				line = lineBuilder.String()
				lineBuilder.Reset()
				if timeoutsBlock := d.nestedBlockHead(line); timeoutsBlock != "" {
					blockName = timeoutsBlock
					continue
				}

				arg := d.parseArgument(line)
				if arg == nil {
					continue
				}
				argName := arg.name
				if blockName != "" {
					argName = fmt.Sprintf("%s.%s", blockName, argName)
				}
				_, ok := r[argName]
				if !ok {
					r[argName] = *arg
				}
			}
		}
	}

	// Check if there was an error scanning the input string
	if err = scanner.Err(); err != nil {
		fmt.Println("Error scanning the string:", err)
	}
	return r, nil
}

func (d Document) beginParse(line string) bool {
	if argumentsHeadlineRegex.MatchString(line) {
		return true
	}
	if timeoutsHeadlineRegex.MatchString(line) {
		return true
	}
	return false
}

func (d Document) endParse(line string) bool {
	return attributesHeadlineRegex.MatchString(line) || importHeadlineRegex.MatchString(line)
}

func (d Document) parseArgument(line string) *argumentDescription {
	line = d.clean(line)
	if !strings.HasPrefix(line, "*") &&
		!strings.Contains(line, "-") {
		return nil
	}
	split := strings.Split(line, " - ")
	if len(split) < 2 {
		return nil
	}
	name := split[0]
	name = strings.TrimPrefix(name, "* `")
	name = strings.TrimSuffix(name, "`")
	description := split[1]
	return &argumentDescription{
		name:         name,
		desc:         description,
		defaultValue: nil,
	}
}

func (d Document) clean(line string) string {
	if line == "" {
		return line
	}
	line = strings.Replace(line, " – ", " - ", -1)
	isDelimiterLine := linq.From([]rune(line)).Distinct().All(func(i interface{}) bool {
		c := i.(rune)
		return c == ' ' || c == '-'
	})
	if isDelimiterLine {
		return line
	}
	if strings.HasPrefix(line, "- ") {
		line = fmt.Sprintf("* %s", line[2:])
	}
	return line
}

func (d Document) nestedBlockHead(line string) string {
	predicts := []func() bool{
		func() bool {
			return strings.HasSuffix(line, "block supports the following:")
		},
		func() bool {
			return backQuoteNameRegexp.MatchString(line) && awsNestedBlockDescriptionKeyWords.All(func(i interface{}) bool {
				return strings.Contains(line, i.(string))
			})
		},
		func() bool {
			return backQuoteNameRegexp.MatchString(line) && strings.HasSuffix(line, "block supports:")
		},
		func() bool {
			return d.isTimeoutsDescription(line)
		},
	}
	for _, p := range predicts {
		if p() {
			nbName := backQuoteNameRegexp.FindString(line)
			if nbName != "" {
				return nbName[1 : len(nbName)-1]
			}
			if d.isTimeoutsDescription(line) {
				return "timeouts"
			}
			return ""
		}
	}
	return ""
}

func (d Document) isTimeoutsDescription(line string) bool {
	return strings.Contains(line, awsTimeoutsUrl) ||
		strings.Contains(line, azureTimeoutsUrl) ||
		strings.Contains(line, googleTimeoutsUrl) ||
		strings.Contains(line, timeoutsUrlSuffix)
}
