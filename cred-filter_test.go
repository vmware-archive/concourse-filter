package main_test

import (
	"os/exec"
	"strings"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredFilter", func() {
	Context("Output to stderr instead", func() {
		It("sends output to stderr instead", func() {
			command := exec.Command(path, "-stderr")
			command.Env = []string{"BORING=boring"}
			command.Stdin = strings.NewReader("boring text")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out.Contents()).To(BeEmpty())
			Expect(string(session.Err.Contents())).To(Equal("[redacted BORING] text"))
		})
	})

	Context("No sensitive credentials available", func() {
		It("outputs as is", func() {
			command := exec.Command(path)
			command.Env = []string{}
			command.Stdin = strings.NewReader("boring text")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(Equal("boring text"))
			Expect(session.Err.Contents()).To(BeEmpty())
		})
	})

	Context("Sensitive credentials available", func() {
		It("filters out those credentials", func() {
			command := exec.Command(path)
			command.Env = []string{"SECRET=secret", "INFO=info"}
			command.Stdin = strings.NewReader("super secret info\nnew line")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(Equal("super [redacted SECRET] [redacted INFO]\nnew line"))
			Expect(session.Err.Contents()).To(BeEmpty())
		})

		Context("sensitive credential env var is whitelisted", func() {
			It("filters out non-white-listed credentials", func() {
				command := exec.Command(path)
				command.Env = []string{"SECRET=secret", "INFO=info", "CREDENTIAL_FILTER_WHITELIST=OTHER1,INFO,OTHER2"}
				command.Stdin = strings.NewReader("super secret info")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(string(session.Out.Contents())).To(Equal("super [redacted SECRET] info"))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})

		Context("the buffer can handle a 256k string", func() {
			It("doesn't crash", func() {
				command := exec.Command(path)
				command.Env = []string{"SECRET=secret", "INFO=info", "CREDENTIAL_FILTER_WHITELIST=OTHER1,INFO,OTHER2"}

				input := string(make([]byte, 256*1024))
				command.Stdin = strings.NewReader(input)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(string(session.Out.Contents())).To(Equal(input))
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})
	})
})
