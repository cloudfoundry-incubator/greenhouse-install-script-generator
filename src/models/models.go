package models

import "errors"

type Release struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type IndexDeployment struct {
	Name     string    `json:"name"`
	Releases []Release `json:"releases"`
}

type ShowDeployment struct {
	Manifest string `json:"manifest"`
}

type Manifest struct {
	Jobs           []Job       `yaml:"jobs"`
	Properties     *Properties `yaml:"properties"`
	InstanceGroups []Job       `yaml:"instance_groups"`
}

func (m *Manifest) FirstRepJob() (*Job, error) {
	jobs := m.Jobs
	if len(jobs) == 0 {
		// 2.0 Manifest
		jobs = m.InstanceGroups
	}

	for _, job := range jobs {
		if job.Properties != nil && job.Properties.Diego != nil && job.Properties.Diego.Rep != nil {
			return &job, nil
		}
	}
	return nil, errors.New("no rep job found")
}

type ConsulProperties struct {
	RequireSSL  *string  `yaml:"require_ssl"`
	CACert      string   `yaml:"ca_cert"`
	AgentCert   string   `yaml:"agent_cert"`
	AgentKey    string   `yaml:"agent_key"`
	EncryptKeys []string `yaml:"encrypt_keys"`
	Agent       struct {
		Domain  string
		Servers struct {
			Lan []string `yaml:"lan"`
		} `yaml:"servers"`
	} `yaml:"agent"`
}

type BBSProperties struct {
	CACert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
	RequireSSL *bool  `yaml:"require_ssl"`
}

type Rep struct {
	Zone       string         `yaml:"zone"`
	BBS        *BBSProperties `yaml:"bbs"`
	RequireTls *bool          `yaml:"require_tls"`
	CACert     string         `yaml:"ca_cert"`
	ServerCert string         `yaml:"server_cert"`
	ServerKey  string         `yaml:"server_key"`
}

type DiegoProperties struct {
	Rep *Rep `yaml:"rep"`
}

type LoggregatorProperties struct {
	Etcd struct {
		Machines []string `yaml:"machines"`
	} `yaml:"etcd"`
	Tls Tls `yaml:"tls"`
}
type Tls struct {
	CA         string `yaml:"ca"`
	CACert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
	Cert       string `yaml:"cert"`
	Key        string `yaml:"key"`
}

type MetronEndpoint struct {
	SharedSecret string `yaml:"shared_secret"`
}

type MetronAgent struct {
	PreferredProtocol *string `yaml:"preferred_protocol"`
	Tls               Tls     `yaml:"tls"`
	TlsClient         Tls     `yaml:"tls_client"`
}

type SyslogProperties struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

type Properties struct {
	Consul              *ConsulProperties      `yaml:"consul"`
	Diego               *DiegoProperties       `yaml:"diego"`
	Loggregator         *LoggregatorProperties `yaml:"loggregator"`
	MetronEndpoint      *MetronEndpoint        `yaml:"metron_endpoint"`
	LoggregatorEndpoint *MetronEndpoint        `yaml:"loggregator_endpoint"`
	MetronAgent         *MetronAgent           `yaml:"metron_agent"`
	Syslog              *SyslogProperties      `yaml:"syslog_daemon_config"`
}

type Job struct {
	Name       string      `yaml:"name"`
	Properties *Properties `yaml:"properties"`
}
