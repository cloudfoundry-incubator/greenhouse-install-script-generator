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
}

func NewInstallerArguments(manifest *Manifest) (*InstallerArguments, error) {
	firstRepJob, err := manifest.FirstRepJob()
	if err != nil {
		return nil, err
	}
	return &InstallerArguments{repJob: firstRepJob, manifest: manifest}, nil
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
