package errorx

func Defer(fn func() error, err *error) {
	*err = Join(*err, fn())
}
