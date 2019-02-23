package out

// ExistingModel is a model that already exists.
type ExistingModel struct {
	Name string
	Enum ExistingEnum
	Int  ExistingInterface
}

// ExistingInput is an input that already exists.
type ExistingInput struct {
	Name string
	Enum ExistingEnum
	Int  ExistingInterface
}

// ExistingEnum is an enum that already exists.
type ExistingEnum string

// ExistingInterface is an interface that already exists.
type ExistingInterface interface {
	IsExistingInterface()
}

// ExistingUnion is a union that already exists.
type ExistingUnion interface {
	IsExistingUnion()
}
