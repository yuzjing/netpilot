package qos

import (
	"fmt"

	"github.com/yuzjing/netpilot/backend/internal/qos/providers"
)

// Manager is responsible for applying and managing QoS rules.
type Manager struct {
	// In the future, we could have dependencies here, like a logger.
}

// NewManager creates a new QoS Manager.
func NewManager() *Manager {
	return &Manager{}
}

// ApplyRule applies a given QoS rule.
func (m *Manager) ApplyRule(rule Rule) error {
	// This is the core logic: dispatching to the correct provider.
	switch rule.Algorithm {
	case CAKE:
		// We extract the specific settings needed for CAKE
		bandwidthMbit, ok := rule.Settings["bandwidth_mbit"].(float64) // JSON numbers are float64
		if !ok {
			return fmt.Errorf("bandwidth_mbit is required for CAKE algorithm and must be a number")
		}
		// and call the specific CAKE provider.
		return providers.ApplyCake(rule.Interface, uint32(bandwidthMbit))
	// case FQ_CODEL:
	//     // Future implementation would go here
	//     return providers.ApplyFqCodel(...)
	default:
		return fmt.Errorf("unsupported QoS algorithm: %s", rule.Algorithm)
	}
}

// DeleteRule removes the QoS rule from a given interface.
func (m *Manager) DeleteRule(ifaceName string) error {
	// For now, we assume deleting is provider-agnostic, but we could
	// add logic here if different algorithms need different cleanup.
	return providers.DeleteRootQdisc(ifaceName)
}

// GetRule fetches the current QoS rule for a given interface. (Not implemented yet)
func (m *Manager) GetRule(ifaceName string) (*Rule, error) {
	// TODO: Implement the logic to read from netlink and construct a Rule.
	return nil, fmt.Errorf("GetRule not implemented yet")
}
