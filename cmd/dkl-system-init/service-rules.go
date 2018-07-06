package main

type ServiceRules struct {
	flags []string
}

func NewServiceRules() *ServiceRules {
	return &ServiceRules{make([]string, 0)}
}

func (r *ServiceRules) Flags(flags ...string) *ServiceRules {
	r.flags = append(r.flags, flags...)
	return r
}

func (r *ServiceRules) Services(serviceNames ...string) *ServiceRules {
	flags := make([]string, len(serviceNames))
	for i, name := range serviceNames {
		flags[i] = "service:" + name
	}
	return r.Flags(flags...)
}

func (r ServiceRules) Check() bool {
	if !services.HasFlag(r.flags...) {
		return false
	}

	return true
}
