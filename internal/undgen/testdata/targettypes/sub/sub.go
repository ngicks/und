package sub

import "github.com/ngicks/und/internal/option"

//undgen:ignore
type Foo struct {
	Yay string
}

func (f Foo) UndPlain() FooPlain {
	return FooPlain{
		Nay: f.Yay,
	}
}

//undgen:ignore
type FooPlain struct {
	Nay string
}

func (f FooPlain) UndRaw() Foo {
	return Foo{
		Yay: f.Nay,
	}
}

//undgen:ignore
type Bar struct {
	O option.Option[string]
}
