package blsgen

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var console consoleItf = &stdConsole{}

// consoleItf define the interface for module level console input and outputs
type consoleItf interface {
	readPassword() (string, error)
	readln() (string, error)
	print(a ...interface{})
	println(a ...interface{})
	printf(format string, a ...interface{})
}

type stdConsole struct{}

func (console *stdConsole) readPassword() (string, error) {
	return console.readln()
}

func (console *stdConsole) readln() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	raw, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

func (console *stdConsole) print(a ...interface{}) {
	fmt.Print(a...)
}

func (console *stdConsole) println(a ...interface{}) {
	fmt.Println(a...)
}

func (console *stdConsole) printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}