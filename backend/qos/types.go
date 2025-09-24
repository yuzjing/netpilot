// Package qos defines the core logic for Quality of Service management.
package qos

// Algorithm represents the supported QoS algorithms.
type Algorithm string

const (
	// CAKE is the "Common Applications Kept Enhanced" algorithm.
	CAKE Algorithm = "cake"
	// FQ_CODEL will be supported in the future.
	// FQ_CODEL Algorithm = "fq_codel"
)

// Settings contains all the parameters for a specific QoS rule.
// We use a flexible map here to easily support different algorithms in the future.
type Settings map[string]interface{}

// Rule represents a complete QoS rule for a network interface.
type Rule struct {
	Interface string    `json:"interface"`
	Algorithm Algorithm `json:"algorithm"`
	Settings  Settings  `json:"settings"`
}
