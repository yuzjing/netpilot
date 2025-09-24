// Package qos is the business logic layer for QoS management.
// It acts as a bridge between the HTTP server and the underlying providers.
package qos

import (
	"fmt"

	"github.com/yuzjing/netpilot/backend/internal/qos/providers"
)

// Manager is responsible for applying and managing QoS rules.
type Manager struct {
	// This struct is a placeholder for any future dependencies, like a logger.
}

// NewManager creates a new QoS Manager.
func NewManager() *Manager {
	return &Manager{}
}

// ApplyRule applies a given QoS rule by dispatching to the correct provider.
func (m *Manager) ApplyRule(rule Rule) error {
	switch rule.Algorithm {
	case CAKE:
		// Extract settings required for the CAKE provider.
		// JSON numbers are decoded as float64, so we must handle that type.
		bandwidthMbitFloat, ok := rule.Settings["bandwidth_mbit"].(float64)
		if !ok {
			return fmt.Errorf("setting 'bandwidth_mbit' is required for CAKE and must be a number")
		}
		// The provider function expects a uint32.
		return providers.ApplyCake(rule.Interface, uint32(bandwidthMbitFloat))

	// Future algorithms can be added here.
	// case FQ_CODEL:
	//     return providers.ApplyFqCodel(...)

	default:
		return fmt.Errorf("unsupported QoS algorithm: %s", rule.Algorithm)
	}
}

// DeleteRule removes the QoS rule from a given interface.
func (m *Manager) DeleteRule(ifaceName string) error {
	// This delegates directly to the provider.
	return providers.DeleteRootQdisc(ifaceName)
}

// GetRule fetches the current QoS rule from the system and assembles it into a Rule struct.
func (m *Manager) GetRule(ifaceName string) (*Rule, error) {
	// 1. Get raw data from the provider. The provider doesn't know about our Rule struct.
	algo, settings, err := providers.GetCurrentRule(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get current rule from provider: %w", err)
	}

	// If the algorithm string is empty, it means no rule was found.
	if algo == "" {
		return nil, nil
	}

	// 2. It is the manager's responsibility to assemble the raw data into the structured Rule.
	rule := &Rule{
		Interface: ifaceName,
		Algorithm: Algorithm(algo),
		Settings:  settings,
	}

	return rule, nil
}
