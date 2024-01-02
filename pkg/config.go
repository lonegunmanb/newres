package pkg

type GenerateMode string

const (
	UniVariable       = "UniVariable"
	MultipleVariables = "MultipleVariables"
)

type Config struct {
	Delimiter string
	Mode      GenerateMode
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
