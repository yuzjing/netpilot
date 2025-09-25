// backend/qos/providers/others.go
package providers

import (
	"fmt"
	"os/exec"
	"strconv"
)

// ApplyTbf applies the Token Bucket Filter qdisc for simple rate limiting.
func ApplyTbf(ifaceName string, bandwidthMbit uint32) error {
	DeleteRootQdisc(ifaceName)
	rate := strconv.FormatUint(uint64(bandwidthMbit), 10) + "mbit"

	// tbf requires a buffer size and limit, we use some sensible defaults.
	// Command: tc qdisc add dev <iface> root tbf rate <rate> buffer 1600 limit 3000
	cmd := exec.Command("tc", "qdisc", "add", "dev", ifaceName, "root", "tbf", "rate", rate, "buffer", "1600", "limit", "3000")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply tbf qdisc: %v, output: %s", err, string(output))
	}
	return nil
}

// ApplyPfifoFast applies the system's default qdisc.
// This is effectively the same as deleting the root qdisc, as the kernel will reinstate pfifo_fast.
func ApplyPfifoFast(ifaceName string) error {
	return DeleteRootQdisc(ifaceName)
}

func ApplySfq(ifaceName string) error {
	DeleteRootQdisc(ifaceName)
	cmd := exec.Command("tc", "qdisc", "add", "dev", ifaceName, "root", "sfq")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply sfq qdisc: %v, output: %s", err, string(output))
	}
	return nil
}
