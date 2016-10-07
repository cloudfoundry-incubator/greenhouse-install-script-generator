package models

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

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

func stringToEncryptKey(str string) string {
	decodedStr, err := base64.StdEncoding.DecodeString(str)
	if err == nil && len(decodedStr) == 16 {
		return str
	}

	key := pbkdf2.Key([]byte(str), nil, 20000, 16, sha1.New)
	return base64.StdEncoding.EncodeToString(key)
}

func (a *InstallerArguments) FillConsul() {
	properties := a.repJob.Properties
	if properties.Consul == nil {
		properties = a.manifest.Properties
	}

	consuls := properties.Consul.Agent.Servers.Lan

	if len(consuls) == 0 {
		fmt.Fprintf(os.Stderr, "Could not find any Consul VMs in your BOSH deployment")
		os.Exit(1)
	}

	a.ConsulIPs = strings.Join(consuls, ",")

	// missing requireSSL implies true
	requireSSL := properties.Consul.RequireSSL
	if requireSSL == nil || *requireSSL != "false" {
		a.ConsulRequireSSL = true
		encryptKey := stringToEncryptKey(properties.Consul.EncryptKeys[0])

		a.Certs["consul_agent.crt"] = properties.Consul.AgentCert
		a.Certs["consul_agent.key"] = properties.Consul.AgentKey
		a.Certs["consul_ca.crt"] = properties.Consul.CACert
		a.Certs["consul_encrypt.key"] = encryptKey
	}

	if properties.Consul.Agent.Domain != "" {
		a.ConsulDomain = properties.Consul.Agent.Domain
	} else {
		a.ConsulDomain = "cf.internal"
	}
}

func (a *InstallerArguments) FillMachineIp(machineIp string) {
	a.MachineIp = machineIp
}

func (a *InstallerArguments) FillBBS() {
	properties := a.repJob.Properties
	if properties.Diego.Rep.BBS == nil {
		properties = a.manifest.Properties
	}

	requireSSL := properties.Diego.Rep.BBS.RequireSSL
	// missing requireSSL implies true
	if requireSSL == nil || *requireSSL {
		a.BbsRequireSsl = true
		a.Certs["bbs_client.crt"] = properties.Diego.Rep.BBS.ClientCert
		a.Certs["bbs_client.key"] = properties.Diego.Rep.BBS.ClientKey
		a.Certs["bbs_ca.crt"] = properties.Diego.Rep.BBS.CACert
	}
}
