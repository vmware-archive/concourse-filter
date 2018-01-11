package main_test

import (
	"bytes"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func runBinary(stdin string, args, env []string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("./cred-filter.exe", args...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

var _ = Describe("CredFilter", func() {
	var args []string
	BeforeEach(func() {
		args = []string{}
	})

	Context("Output to stderr instead", func() {
		It("sends output to stderr instead", func() {
			args = []string{"-stderr"}
			output, stderr, err := runBinary("boring text", args, []string{})
			Expect(err).To(BeNil())
			Expect(output).To(Equal(""))
			Expect(stderr).To(Equal("boring text\n"))
		})
	})

	Context("No sensitive credentials available", func() {
		It("outputs as is", func() {
			env := []string{}
			output, stderr, err := runBinary("boring text", args, env)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("boring text\n"))
			Expect(stderr).To(Equal(""))
		})
	})
	Context("Sensitive credentials available", func() {
		It("filters out those credentials", func() {
			env := []string{"SECRET=secret", "INFO=info"}
			output, _, err := runBinary("super secret info\nnew line", args, env)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("super [redacted SECRET] [redacted INFO]\nnew line\n"))
		})
		Context("sensitive credential env var is whitelisted", func() {
			It("filters out non-white-listed credentials", func() {
				env := []string{"SECRET=secret", "INFO=info", "CREDENTIAL_FILTER_WHITELIST=OTHER1,INFO,OTHER2"}
				output, _, err := runBinary("super secret info", args, env)
				Expect(err).To(BeNil())
				Expect(output).To(Equal("super [redacted SECRET] info\n"))
			})
		})
		Context("the buffer can handle a 256k string", func() {
			It("doesn't crash", func() {
				env := []string{"SECRET=secret", "INFO=info", "CREDENTIAL_FILTER_WHITELIST=OTHER1,INFO,OTHER2"}
				input := make([]byte, 256*1024)

				_, _, err := runBinary(string(input[:]), args, env)
				Expect(err).To(BeNil())
			})
		})
	})
})
