package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestConcourseFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConcourseFilter Suite")
}

var path string

var _ = BeforeSuite(func() {
	var err error
	path, err = gexec.Build("github.com/pivotal-cf-experimental/concourse-filter")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
