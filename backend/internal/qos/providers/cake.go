package providers

import (
	"fmt"
	"os/exec"
	"strconv"
)

// ApplyCake applies the CAKE qdisc by executing the 'tc' command.
func ApplyCake(ifaceName string, bandwidthMbit uint32) error {
	// First, try to delete any existing root qdisc to avoid conflicts.
	DeleteRootQdisc(ifaceName)

	// Convert bandwidth to a string like "100mbit"
	rate := strconv.FormatUint(uint64(bandwidthMbit), 10) + "mbit"

	// Prepare the command: sudo tc qdisc add dev <iface> root cake bandwidth <rate>
	cmd := exec.Command("sudo", "tc", "qdisc", "add", "dev", ifaceName, "root", "cake", "bandwidth", rate)

	// Execute the command and capture its output (stdout and stderr combined)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If there's an error, return a detailed message including the command's output
		return fmt.Errorf("failed to apply cake qdisc: %v, output: %s", err, string(output))
	}

	return nil
}

// DeleteRootQdisc removes the root qdisc by executing the 'tc' command.
func DeleteRootQdisc(ifaceName string) error {
	cmd := exec.Command("sudo", "tc", "qdisc", "del", "dev", ifaceName, "root")

	// We can ignore errors here, as the qdisc might not exist.
	// In a production app, we might want to log this.
	cmd.Run()

	return nil
}
