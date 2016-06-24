package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//newEnvStringReplacer creates a string replacer for env variable text
func newEnvStringReplacer() *strings.Replacer {

	var envVars []string

	for _, envVar := range os.Environ() {
		pair := strings.Split(envVar, "=")
		if pair[1] != "" {
			envVars = append(envVars, pair[1])
			envVars = append(envVars, "[redacted]")
		}
	}

	return strings.NewReplacer(envVars...)
}

func main() {

	envStringReplacer := newEnvStringReplacer()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(envStringReplacer.Replace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
