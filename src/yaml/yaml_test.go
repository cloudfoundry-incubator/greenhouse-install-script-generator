package yaml_test

import (
	"yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Manifest struct {
	Properties *struct {
		Syslog *struct{} `yaml:"syslog_daemon_config"`
	} `yaml:"properties"`
}

type Properties struct {
}

var _ = Describe("Yaml", func() {
	It("works with empty objects", func() {
		// y, err := ioutil.ReadFile("../integration/syslog_with_empty_config.yml")
		// Expect(err).ToNot(HaveOccurred())
		y := `
properties:
  syslog_daemon_config:
`
		err := yaml.Unmarshal([]byte(y), &Manifest{})
		Expect(err).To(BeNil())
	})
})
