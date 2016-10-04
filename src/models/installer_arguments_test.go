package models_test

import (
	. "models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstallerArguments", func() {
	var (
		repJob   Job
		manifest Manifest
	)
	BeforeEach(func() {
		repJob = Job{
			Properties: &Properties{
				Diego: &DiegoProperties{
					Rep: &Rep{},
				},
			},
		}
		manifest = Manifest{
			Jobs: []Job{repJob},
			Properties: &Properties{
				MetronAgent: &MetronAgent{},
				Loggregator: &LoggregatorProperties{},
			},
		}
	})

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
		const sharedSecret = "foo"

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

	Describe("FillMetronAgent", func() {
		It("does not copy certs when TLS is not the preferred protocol", func() {
			tcp := "tcp"
			manifest.Properties.MetronAgent.PreferredProtocol = &tcp

			args, err := NewInstallerArguments(&manifest)
			Expect(err).To(BeNil())

			args.FillMetronAgent()
			Expect(args.Certs).To(BeEmpty())
			Expect(args.MetronPreferTLS).To(BeFalse())
		})

		Context("new style loggregator properties", func() {
			It("copies certs from the manifest when TLS is enabled", func() {
				tls := "tls"
				manifest.Properties.MetronAgent.PreferredProtocol = &tls
				manifest.Properties.Loggregator.Tls = Tls{CACert: "cacert"}
				manifest.Properties.MetronAgent.Tls = Tls{
					ClientKey:  "clientkey",
					ClientCert: "clientcert",
				}

				args, err := NewInstallerArguments(&manifest)
				Expect(err).To(BeNil())

				args.FillMetronAgent()
				Expect(args.Certs["metron_agent.crt"]).To(Equal("clientcert"))
				Expect(args.Certs["metron_agent.key"]).To(Equal("clientkey"))
				Expect(args.Certs["metron_ca.crt"]).To(Equal("cacert"))
				Expect(args.MetronPreferTLS).To(BeTrue())
			})
		})

		Context("old style loggregator properties", func() {
			It("copies certs from the manifest when TLS is enabled", func() {
				tls := "tls"
				manifest.Properties.MetronAgent.PreferredProtocol = &tls
				manifest.Properties.Loggregator.Tls = Tls{CA: "cacert"}
				manifest.Properties.MetronAgent.TlsClient = Tls{
					Key:  "clientkey",
					Cert: "clientcert",
				}

				args, err := NewInstallerArguments(&manifest)
				Expect(err).To(BeNil())

				args.FillMetronAgent()
				Expect(args.Certs["metron_agent.crt"]).To(Equal("clientcert"))
				Expect(args.Certs["metron_agent.key"]).To(Equal("clientkey"))
				Expect(args.Certs["metron_ca.crt"]).To(Equal("cacert"))
				Expect(args.MetronPreferTLS).To(BeTrue())
			})
		})
	})
})
