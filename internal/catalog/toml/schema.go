package toml

// catalogFile mirrors the repository-local TOML schema. It stays private to the
// adapter so planning remains format-agnostic.
type catalogFile struct {
	Schema   string          `toml:"schema"`
	Version  int             `toml:"version"`
	Tools    []resourceEntry `toml:"tools"`
	Runtimes []resourceEntry `toml:"runtimes"`
	Packages []resourceEntry `toml:"packages"`
	Dotfiles []resourceEntry `toml:"dotfiles"`
	Bundles  []bundleEntry   `toml:"bundles"`
	Profiles []profileEntry  `toml:"profiles"`
}

type resourceEntry struct {
	ID             string   `toml:"id"`
	Description    string   `toml:"description"`
	DependsOn      []string `toml:"depends_on"`
	ConfigRequired []string `toml:"config_required"`
	OS             []string `toml:"os"`
	Arch           []string `toml:"arch"`
	Distro         []string `toml:"distro"`
	WSL            *bool    `toml:"wsl"`
}

type bundleEntry struct {
	ID        string   `toml:"id"`
	Resources []string `toml:"resources"`
}

type profileEntry struct {
	ID        string   `toml:"id"`
	Bundles   []string `toml:"bundles"`
	Resources []string `toml:"resources"`
}
