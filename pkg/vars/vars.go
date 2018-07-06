package vars

import (
	"bytes"

	"novit.nc/direktil/pkg/config"
)

type Var struct {
	Template []byte
	Value    []byte
}

func Vars(cfg *config.Config) []Var {
	res := make([]Var, 0)

	for _, arg := range BootArgs() {
		if !bytes.HasPrefix(arg, bootVarPrefix) {
			continue
		}

		parts := bytes.SplitN(arg[len(bootVarPrefix):], []byte{'='}, 2)

		res = append(res, Var{
			Template: append(append([]byte("$(var:"), parts[0]...), ')'),
			Value:    parts[1],
		})
	}

configVarsLoop:
	for _, v := range cfg.Vars {
		t := []byte("$(var:" + v.Name + ")")
		for _, prev := range res {
			if bytes.Equal(prev.Template, t) {
				continue configVarsLoop
			}
		}

		res = append(res, Var{t, []byte(v.Default)})
	}

	return res
}

// Substitute variables in src
func Substitute(src []byte, cfg *config.Config) (dst []byte) {
	dst = src

	for _, bv := range Vars(cfg) {
		if !bytes.Contains(dst, bv.Template) {
			continue
		}

		v := bytes.TrimSpace(bv.Value)

		dst = bytes.Replace(dst, bv.Template, v, -1)
	}

	return
}
