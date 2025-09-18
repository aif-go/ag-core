package types

import (
	"log/slog"
)

var (
	globalDependencies = map[string]string{}
)

// AddGlobalDependency adds dependency for all generators
func AddGlobalDependency(ref, path string) bool {
	if _, ok := globalDependencies[ref]; !ok {
		globalDependencies[ref] = path
		return true
	}
	return false
}

// AddGlobalDependencys adds dependency for all generators
func AddGlobalDependencys(defs ...string) bool {
	if len(defs)%2 != 0 {
		slog.Error("AddGlobalDependencys: defs must be even")
		return false
	}
	for i := 0; i < len(defs); i += 2 {
		if !AddGlobalDependency(defs[i], defs[i+1]) {
			return false
		}
	}
	return true
}
