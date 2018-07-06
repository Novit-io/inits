package main

/*
func init() {
	legacyService := func(name string, needs ...string) *Service {
		return &Service{
			Name:    name,
			PreCond: NewServiceRules().Services(needs...).Check,
			Run: func(notify func()) bool {
				return Exec(func() {}, "/etc/init.d/"+name, "start", "--nodeps")
			},
		}
	}

    services.Register([]*Service{
        legacyService("modules-load"),
        legacyService("modules", "modules-load"),
        legacyService("dmesg", "udev", "modules"),
    }...)
}
// */
