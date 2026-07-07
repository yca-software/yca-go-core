package chi_server

import (
	"golang.org/x/sync/errgroup"
)

// CleanupDependency is a resource that must be released on shutdown (e.g. DB pools, Redis clients).
type CleanupDependency interface {
	Cleanup()
}

// CleanupDependenciesConcurrently runs Cleanup on each dependency in parallel.
func CleanupDependenciesConcurrently(deps []CleanupDependency) {
	var g errgroup.Group
	for _, d := range deps {
		d := d
		g.Go(func() error {
			d.Cleanup()
			return nil
		})
	}
	_ = g.Wait()
}
