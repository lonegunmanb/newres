package pkg

type postProcessor interface {
	action(r *resourceBlock)
}
