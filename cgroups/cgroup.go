package cgroups

const (
	defaultPerm = 0755
)

type Resource struct {
	Memory *Memory `json:"memory,omitempty"` // Memory restriction configuration
	CPU    *CPU    `json:"cpu,omitempty"`    // CPU resource restriction configuration
	Pids   *Pids   `json:"pids,omitempty"`   // Task resource restriction configuration.
}

type Cgroup struct {
	path       string // cgroup relative path
	subSystems []SubSystem
}

type SubSystem interface {
	create(path string) error
	delete(path string) error
	add(path string, pid int) error // add process to cgroup
}

type cgroupFile struct {
	name    string // file name
	content string // file content
}

// new cgroup
func New(path string, resource Resource) (*Cgroup, error) {
	cgroup := &Cgroup{
		path:       path,
		subSystems: []SubSystem{},
	}

	if resource.Memory != nil {
		cgroup.subSystems = append(cgroup.subSystems, resource.Memory)
	}
	if resource.CPU != nil {
		cgroup.subSystems = append(cgroup.subSystems, resource.CPU)
	}
	if resource.Pids != nil {
		cgroup.subSystems = append(cgroup.subSystems, resource.Pids)
	}

	if err := cgroup.Create(); err != nil {
		return nil, err
	}

	return cgroup, nil
}

func (cg Cgroup) Create() error {
	for _, subSystem := range cg.subSystems {
		if err := subSystem.create(cg.path); err != nil {
			cg.Delete() // delete created cgroup
			return err
		}
	}

	return nil
}

func (cg Cgroup) Add(pid int) error {
	for _, subSystem := range cg.subSystems {
		if err := subSystem.add(cg.path, pid); err != nil {
			return err
		}
	}

	return nil
}

func (cg Cgroup) Delete() error {
	for _, subSystem := range cg.subSystems {
		if err := subSystem.delete(cg.path); err != nil {
			return err
		}
	}

	return nil
}
