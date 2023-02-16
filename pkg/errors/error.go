package errors

type Error interface {
	Error() string
}

// Todo : Maybe Change this location //Configure error mechanism
type ErrorDefinition struct {
	number   int
	prefix   string
	template string
}
