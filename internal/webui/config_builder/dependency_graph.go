package config_builder

// DependencyGraph represents the dependency graph for the builder
// This is a minimal scaffold for now

type DependencyGraph struct {
	Nodes []string
	Edges []struct {
		From string
		To   string
	}
}
