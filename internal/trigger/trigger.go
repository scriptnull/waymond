package trigger

type Type string

type Interface interface {
	Type() Type
	Register() error
}
