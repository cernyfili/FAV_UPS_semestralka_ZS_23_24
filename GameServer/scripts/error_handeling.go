package main

import (
	"fmt"
	"github.com/pkg/errors"
)

func someFunction() error {
	return errors.Wrap(errors.New("Some error"), "someFunction")
}

func someFunction2() error {
	return errors.Wrap(someFunction(), "someFunction2")
}

func someFunction3() error {
	return errors.Wrap(someFunction2(), "someFunction3")
}

func main() {
	if err := someFunction3(); err != nil {
		fmt.Printf("Error: %v Stack trace: %+v ", err, err) // %+v prints the stack trace
		panic(err)
	}
}
