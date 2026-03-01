package resolver

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// ResolveConstraint returns the highest version in available that satisfies constraint.
// Constraint may be exact (1.0.0), caret (^1.0.0), or tilde (~1.0.0).
// If constraint is exact and in available, returns it; otherwise uses semver range.
func ResolveConstraint(constraint string, available []string) (string, error) {
	if len(available) == 0 {
		return "", fmt.Errorf("no versions available")
	}
	// Try as exact first
	for _, v := range available {
		if v == constraint {
			return v, nil
		}
	}
	// Parse as semver constraint
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		// Not a range; treat as exact and fail if not in list
		return "", fmt.Errorf("invalid version constraint %q: %w", constraint, err)
	}
	var best *semver.Version
	for _, s := range available {
		ver, err := semver.NewVersion(s)
		if err != nil {
			continue
		}
		if c.Check(ver) && (best == nil || ver.GreaterThan(best)) {
			best = ver
		}
	}
	if best == nil {
		return "", fmt.Errorf("no version satisfies %q (available: %v)", constraint, available)
	}
	return best.Original(), nil
}
