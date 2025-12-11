package pkg

type GenerateMode string

const (
	UniVariable       = "UniVariable"
	MultipleVariables = "MultipleVariables"
)

type Config struct {
	Delimiter string
	Mode      GenerateMode
	// VariablePrefix overrides the default variable name prefix (resource type without vendor).
	// If VariablePrefixSet is false, the default will be used. If true, the provided value is used even if empty.
	VariablePrefix    string
	VariablePrefixSet bool
}

func (c Config) GetDelimiter() string {
	if c.Delimiter == "" {
		return "EOT"
	}
	return c.Delimiter
}

func (c Config) GetMode() GenerateMode {
	if c.Mode == "" {
		return MultipleVariables
	}
	return c.Mode
}

func (c Config) GetVariablePrefix(defaultPrefix string) string {
	if c.VariablePrefixSet {
		// honor explicit value, including empty string
		return c.VariablePrefix
	}
	if c.VariablePrefix != "" {
		// backward compatibility: if caller set VariablePrefix but didn't mark as set
		return c.VariablePrefix
	}
	return defaultPrefix
}
