package kubernetes

import "fmt"

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("Not found: %s", string(e))
}

type ErrNotTraversable string

func (e ErrNotTraversable) Error() string {
	return fmt.Sprintf("Not a traversable type: %s", string(e))
}

type ErrMissingArrayTraversal string

func (e ErrMissingArrayTraversal) Error() string {
	return fmt.Sprintf("Array traversal indicator missing: %s", string(e))
}

type ErrNotAnArray string

func (e ErrNotAnArray) Error() string {
	return fmt.Sprintf("Not an array: %s", string(e))
}

type ErrNotKeyValue string

func (e ErrNotKeyValue) Error() string {
	return fmt.Sprintf("Path item is not a key/value object: %s", string(e))
}

type ErrIncorrectType string

func (e ErrIncorrectType) Error() string {
	return fmt.Sprintf("Incoorect type: %s", string(e))
}

type ErrIndexNotation string

func (e ErrIndexNotation) Error() string {
	return fmt.Sprintf("Cannot append to array using index notation")
}
