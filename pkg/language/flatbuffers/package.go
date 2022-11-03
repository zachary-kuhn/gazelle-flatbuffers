package flatbuffers

type Package struct {
	Name        string
	RuleName    string
	Files       map[string]FileInfo
	Imports     map[string]bool
	Options     map[string]string
	HasServices bool
}

func newPackage(name string) *Package {
	return &Package{
		Name:    name,
		Files:   map[string]FileInfo{},
		Imports: map[string]bool{},
		Options: map[string]string{},
	}
}

func (p *Package) addFile(info FileInfo) {
	p.Files[info.Name] = info
	for _, imp := range info.Imports {
		p.Imports[imp] = true
	}
	p.HasServices = p.HasServices || info.HasServices
}
