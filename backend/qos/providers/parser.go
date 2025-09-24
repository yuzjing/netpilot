package providers

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	// 看！我们删除了 "github.com/yuzjing/netpilot/backend/internal/qos" 的导入
)

var tcOutputRegex = regexp.MustCompile(`qdisc cake .* bandwidth (\d+[KMG]?bit)`)

// GetCurrentRule parses tc output and returns the raw findings.
// It returns: (algorithmName, settingsMap, error)
// It no longer knows about the qos.Rule struct. This breaks the cycle.
func GetCurrentRule(ifaceName string) (string, map[string]interface{}, error) {
	cmd := exec.Command("tc", "qdisc", "show", "dev", ifaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "does not exist") {
			return "", nil, nil // No rule found, return nil without error
		}
		return "", nil, fmt.Errorf("failed to execute tc command: %v, output: %s", err, string(output))
	}

	matches := tcOutputRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", nil, nil // No CAKE rule found
	}

	// The provider's job is to provide raw data, not a formatted struct.
	algorithm := "cake"
	settings := map[string]interface{}{
		"bandwidth": matches[1], // e.g., "500Mbit"
	}

	return algorithm, settings, nil
}
