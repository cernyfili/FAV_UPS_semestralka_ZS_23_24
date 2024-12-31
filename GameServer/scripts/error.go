package main

/*
import (
	"fmt"
	"runtime/debug"
)

func someFunction() error {
	return fmt.Errorf("Some error")
}

func someFunction2() error {
	return fmt.Errorf("Some error 2: %w", someFunction())
}

func someFunction3() error {
	return fmt.Errorf("Some error 3: %w", someFunction2())
}

func main() {
	if err := someFunction3(); err != nil {
		fmt.Println("Error:", err)

		// Capture and print stack trace
		fmt.Println("Triggering panic with stack trace:")
		panicWithStack(err)
	}
}

func panicWithStack(err error) {
	stack := debug.Stack()
	panic(fmt.Sprintf("%v\nStack trace:\n%s", err, string(stack)))
}
*/
