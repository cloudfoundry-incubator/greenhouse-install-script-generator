package models

type InstallerArguments struct {
	repJob           *Job
	manifest         *Manifest
	ConsulRequireSSL bool
	ConsulIPs        string
	EtcdCluster      string
	Zone             string
	SharedSecret     string
	Username         string
	Password         string
	SyslogHostIP     string
	SyslogPort       string
	BbsRequireSsl    bool
	MachineIp        string
	MetronPreferTLS  bool
	ConsulDomain     string
	Certs            map[string]string
}

func NewInstallerArguments(manifest *Manifest) (*InstallerArguments, error) {
	firstRepJob, err := manifest.FirstRepJob()
	if err != nil {
		return nil, err
	}
	return &InstallerArguments{
		repJob:   firstRepJob,
		manifest: manifest,
		Certs:    make(map[string]string),
	}, nil
}

func (a *InstallerArguments) FillEtcdCluster() {
	properties := a.repJob.Properties
	if properties.Loggregator == nil {
		properties = a.manifest.Properties
	}

	a.EtcdCluster = properties.Loggregator.Etcd.Machines[0]
}

func (a *InstallerArguments) FillSharedSecret() {
	properties := a.repJob.Properties
	if properties.MetronEndpoint == nil && properties.LoggregatorEndpoint == nil {
		properties = a.manifest.Properties
	}
	if properties.MetronEndpoint != nil {
		a.SharedSecret = properties.MetronEndpoint.SharedSecret
	} else if properties.LoggregatorEndpoint != nil {
		a.SharedSecret = properties.LoggregatorEndpoint.SharedSecret
	}
}

func (a *InstallerArguments) FillMetronAgent() {
	properties := a.repJob.Properties

	if properties.MetronAgent == nil || properties.MetronAgent.PreferredProtocol == nil {
		properties = a.manifest.Properties
	}

	if properties != nil && properties.MetronAgent != nil && properties.MetronAgent.PreferredProtocol != nil {
		if *properties.MetronAgent.PreferredProtocol == "tls" {
			a.MetronPreferTLS = true
			if properties.Loggregator.Tls.CACert != "" {
				a.Certs["metron_agent.crt"] = properties.MetronAgent.Tls.ClientCert
				a.Certs["metron_agent.key"] = properties.MetronAgent.Tls.ClientKey
				a.Certs["metron_ca.crt"] = properties.Loggregator.Tls.CACert
			} else {
				a.Certs["metron_agent.crt"] = properties.MetronAgent.TlsClient.Cert
				a.Certs["metron_agent.key"] = properties.MetronAgent.TlsClient.Key
				a.Certs["metron_ca.crt"] = properties.Loggregator.Tls.CA
			}
		}
	}
}

func (a *InstallerArguments) FillSyslog() {
	properties := a.repJob.Properties
	if properties.Syslog == nil && a.manifest.Properties != nil {
		properties = a.manifest.Properties
	}

	if properties.Syslog == nil {
		return
	}

	a.SyslogHostIP = properties.Syslog.Address
	a.SyslogPort = properties.Syslog.Port
}
