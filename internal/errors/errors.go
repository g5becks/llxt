package errs

import "github.com/samber/oops"

//nolint:gochecknoglobals // error builders are intentionally global for convenience
var (
	// RegistryErr creates errors in the registry domain.
	RegistryErr = oops.In(DomainRegistry).Tags("registry")
	// HTTPErr creates errors in the HTTP domain.
	HTTPErr = oops.In(DomainHTTP).Tags("http", "network")
	// ConfigErr creates errors in the config domain.
	ConfigErr = oops.In(DomainConfig).Tags("config")
	// GitHubErr creates errors in the GitHub domain.
	GitHubErr = oops.In(DomainGitHub).Tags("github", "api")
	// GeneratorErr creates errors in the generator domain.
	GeneratorErr = oops.In(DomainGenerator).Tags("generator")
)
