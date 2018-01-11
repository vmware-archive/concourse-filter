package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func whiteList() map[string]bool {
	r, _ := regexp.Compile("^CREDENTIAL_FILTER_WHITELIST=")
	whiteListMap := map[string]bool{}
	for _, envVar := range os.Environ() {
		if r.MatchString(envVar) {
			pair := strings.Split(envVar, "=")
			envVarWhitelist := pair[1]
			for _, key := range strings.Split(envVarWhitelist, ",") {
				whiteListMap[key] = true
			}
		}
	}
	return whiteListMap
}

//newEnvStringReplacer creates a string replacer for env variable text
func newEnvStringReplacer() *strings.Replacer {
	var envVars []string

	whiteList := whiteList()

	for _, envVar := range os.Environ() {
		pair := strings.Split(envVar, "=")
		envVarName := pair[0]
		envVarValue := pair[1]
		if !whiteList[envVarName] && envVarValue != "" {
			envVars = append(envVars, envVarValue)
			redactedOutput := "[redacted " + envVarName + "]"
			envVars = append(envVars, redactedOutput)
		}
	}

	return strings.NewReplacer(envVars...)
}

func main() {
	envStringReplacer := newEnvStringReplacer()
	buffer := make([]byte, 257*1024)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(buffer, 257*1024)

	for scanner.Scan() {
		output := os.Stdout
		if len(os.Args) > 1 && os.Args[1] == "-stderr" {
			output = os.Stderr
		}
		fmt.Fprintln(output, envStringReplacer.Replace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
