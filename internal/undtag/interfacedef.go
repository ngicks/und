package undtag

type ElasticLike interface {
	UndLike
	Len() int
	HasNull() bool
}

type UndLike interface {
	IsDefined() bool
	IsNull() bool
	IsUndefined() bool
}

type OptionLike interface {
	IsNone() bool
	IsSome() bool
}
