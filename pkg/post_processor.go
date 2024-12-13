package pkg

type postProcessor interface {
	action(terraformConfig string, cfg Config) (string, error)
}
