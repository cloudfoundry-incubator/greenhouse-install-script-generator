package models_test

import (
	. "models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstallerArguments", func() {
	Describe("NewInstallerArguments", func() {
		It("errors when there are no rep jobs in the manifest", func() {
			manifest := Manifest{
				Jobs: []Job{},
			}
			_, err := NewInstallerArguments(&manifest)
			Expect(err).To(HaveOccurred())
		})
	})
})
