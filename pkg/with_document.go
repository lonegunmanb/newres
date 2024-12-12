package pkg

type withDocument interface {
	Doc() (map[string]argumentDescription, error)
}
