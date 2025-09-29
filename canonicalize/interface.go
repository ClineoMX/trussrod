package canonicalize

type Canonical interface {
	Canonicalize() ([]byte, error)
}
