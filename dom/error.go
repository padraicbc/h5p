package dom

// SelectorError wraps selector parsing and matching failures with a friendly
// message that can be surfaced to callers.
type SelectorError struct {
	Msg string
}

// Error implements the error interface.
func (e SelectorError) Error() string {
	return e.Msg
}
