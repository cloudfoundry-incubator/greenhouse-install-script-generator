package integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var generatePath string

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(5 * time.Second)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	var err error
	generatePath, err = gexec.Build("generate")
	Expect(err).NotTo(HaveOccurred())
})
