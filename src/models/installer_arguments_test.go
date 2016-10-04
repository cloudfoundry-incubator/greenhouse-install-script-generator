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

	Describe("FillSharedSecret", func() {
		var (
			repJob   Job
			manifest Manifest
		)
		const sharedSecret = "foo"

		BeforeEach(func() {
			repJob = Job{
				Properties: &Properties{
					Diego: &DiegoProperties{
						Rep: &Rep{},
					},
				},
			}
			manifest = Manifest{
				Jobs:       []Job{repJob},
				Properties: &Properties{},
			}
		})

		It("works when the job uses the legacy loggregator_endpoint property", func() {
			repJob.Properties.LoggregatorEndpoint = &MetronEndpoint{
				SharedSecret: sharedSecret,
			}

			args, err := NewInstallerArguments(&manifest)
			Expect(err).To(BeNil())

			args.FillSharedSecret()
			Expect(args.SharedSecret).To(Equal(sharedSecret))
		})

		It("works when the job uses the new metron_endpoint property", func() {
			repJob.Properties.MetronEndpoint = &MetronEndpoint{
				SharedSecret: sharedSecret,
			}

			args, err := NewInstallerArguments(&manifest)
			Expect(err).To(BeNil())

			args.FillSharedSecret()
			Expect(args.SharedSecret).To(Equal(sharedSecret))
		})

		It("works when the global properties use the legacy loggregator_endpoint property", func() {
			manifest.Properties.LoggregatorEndpoint = &MetronEndpoint{SharedSecret: sharedSecret}

			args, err := NewInstallerArguments(&manifest)
			Expect(err).To(BeNil())

			args.FillSharedSecret()
			Expect(args.SharedSecret).To(Equal(sharedSecret))
		})

		It("works when the global properties use the new metron_endpoint property", func() {
			manifest.Properties.MetronEndpoint = &MetronEndpoint{SharedSecret: sharedSecret}

			args, err := NewInstallerArguments(&manifest)
			Expect(err).To(BeNil())

			args.FillSharedSecret()
			Expect(args.SharedSecret).To(Equal(sharedSecret))
		})
	})
})
