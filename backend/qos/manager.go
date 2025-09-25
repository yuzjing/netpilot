// Package qos is the business logic layer for QoS management.
package qos

import (
	"fmt"

	"github.com/yuzjing/netpilot/backend/qos/providers"
)

// Manager is responsible for applying and managing QoS rules.
type Manager struct{}

// NewManager creates a new QoS Manager.
func NewManager() *Manager { return &Manager{} }

// ApplyRule applies a given QoS rule by dispatching to the correct provider.
func (m *Manager) ApplyRule(rule Rule) error {
	// This switch correctly handles dispatching based on the Algorithm type.
	switch rule.Algorithm {
	case CAKE:
		bandwidthMbitFloat, ok := rule.Settings["bandwidth_mbit"].(float64)
		if !ok {
			return fmt.Errorf("setting 'bandwidth_mbit' is required for CAKE")
		}
		return providers.ApplyCake(rule.Interface, uint32(bandwidthMbitFloat))

	case FQ_CODEL:
		return providers.ApplyFqCodel(rule.Interface)

	case TBF:
		bandwidthMbitFloat, ok := rule.Settings["bandwidth_mbit"].(float64)
		if !ok {
			return fmt.Errorf("setting 'bandwidth_mbit' is required for TBF")
		}
		return providers.ApplyTbf(rule.Interface, uint32(bandwidthMbitFloat))

	case SFQ:
		return providers.ApplySfq(rule.Interface)

	case PFIFO_FAST:
		return providers.ApplyPfifoFast(rule.Interface)

	default:
		return fmt.Errorf("unsupported QoS algorithm: %s", rule.Algorithm)
	}
}

// DeleteRule removes the QoS rule from a given interface.
func (m *Manager) DeleteRule(ifaceName string) error {
	return providers.DeleteRootQdisc(ifaceName)
}

// GetRule fetches and assembles the current QoS rule.
// 【核心修正】This function now correctly calls the new `GetCurrentRule` function.
func (m *Manager) GetRule(ifaceName string) (*Rule, error) {
	// 1. Call the new, smart parser from the providers package.
	algo, settings, err := providers.GetCurrentRule(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get current rule from provider: %w", err)
	}

	// If the algorithm string is empty, it means no recognizable rule was found.
	if algo == "" {
		return nil, nil
	}

	// 2. The manager assembles the raw data into the structured Rule.
	rule := &Rule{
		Interface: ifaceName,
		Algorithm: Algorithm(algo),
		Settings:  settings,
	}

	return rule, nil
}
