package out_port

import "time"

// BuildRunner executes the site build. argv is an argv array; implementations
// must never hand it to a shell. env is KEY=VALUE entries merged over the
// inherited process environment (entries rather than a map: the config loader
// lowercases map keys, which would corrupt names like ARGSEA_API_URL). The
// returned output is the combined stdout+stderr, returned even when the
// command fails so the caller can surface it.
type BuildRunner interface {
	Run(dir string, argv []string, env []string, timeout time.Duration) (string, error)
}
