package models

type Manifest struct {
	Jobs           []Job       `yaml:"jobs"`
	Properties     *Properties `yaml:"properties"`
	InstanceGroups []Job       `yaml:"instance_groups"`
}

func (m *Manifest) firstRepJob() Job {
	jobs := m.Jobs
	if len(jobs) == 0 {
		// 2.0 Manifest
		jobs = m.InstanceGroups
	}

	for _, job := range jobs {
		if job.Properties.Diego != nil && job.Properties.Diego.Rep != nil {
			return job
		}

	}
	panic("no rep jobs found")
}

func (m *Manifest) FillEtcdCluster(args *InstallerArguments) error {
	repJob := m.firstRepJob()
	properties := repJob.Properties
	if properties.Loggregator == nil {
		properties = m.Properties
	}

	args.EtcdCluster = properties.Loggregator.Etcd.Machines[0]
	return nil
}
