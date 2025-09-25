// backend/qos/providers/fqcodel.go
package providers

import (
	"fmt"
	"os/exec"
)

// ApplyFqCodel applies the fq_codel qdisc.
// NOTE: fq_codel itself doesn't have a bandwidth parameter.
// For real-world use, it's often paired with another qdisc like HTB for rate limiting.
// For this MVP, we will apply it directly, which is useful for solving bufferbloat
// on links that are already shaped by the ISP.
func ApplyFqCodel(ifaceName string) error {
	// First, clean up any existing root qdisc.
	DeleteRootQdisc(ifaceName)

	// Prepare the command: sudo tc qdisc add dev <iface> root fq_codel
	cmd := exec.Command("tc", "qdisc", "add", "dev", ifaceName, "root", "fq_codel")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply fq_codel qdisc: %v, output: %s", err, string(output))
	}

	return nil
}
