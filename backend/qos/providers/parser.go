package providers

import (
	"fmt"
	"os/exec"
	"regexp"

	"strings"
)

// We define a generic regex to find the primary algorithm.
var qdiscKindRegex = regexp.MustCompile(`qdisc (\w+)`)

// GetCurrentRule has been rewritten to be a robust, multi-stage parser.
func GetCurrentRule(ifaceName string) (string, map[string]interface{}, error) {
	cmd := exec.Command("tc", "-s", "qdisc", "show", "dev", ifaceName)
	output, err := cmd.CombinedOutput()
	rawOutput := string(output)

	if err != nil {
		if strings.Contains(rawOutput, "does not exist") {
			return "", nil, nil
		}
		return "", nil, fmt.Errorf("failed to execute tc command: %v, output: %s", err, rawOutput)
	}

	// Step 1: Find the primary algorithm name. This is our anchor.
	kindMatches := qdiscKindRegex.FindStringSubmatch(rawOutput)
	if len(kindMatches) < 2 {
		// This can happen on some systems where the default qdisc is not shown.
		// We can infer it's likely a system default.
		if strings.TrimSpace(rawOutput) == "" {
			return "pfifo_fast", make(map[string]interface{}), nil
		}
		return "unknown", map[string]interface{}{"raw_output": rawOutput}, nil
	}
	algorithm := kindMatches[1]
	settings := make(map[string]interface{})

	// Step 2: Based on the found algorithm, try to parse its specific parameters.
	// This makes our parsing robust and handles optional parameters.
	switch algorithm {
	case "cake":
		// Cake's format is consistent
		if matches := regexp.MustCompile(`bandwidth (\d+[KMG]?bit)(.*)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["bandwidth"] = matches[1]
			settings["options"] = strings.TrimSpace(matches[2])
		}
	case "tbf":
		// TBF has required and optional parts
		if matches := regexp.MustCompile(`rate (\d+[KkMmGg]?bit)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["rate"] = matches[1]
		}
		if matches := regexp.MustCompile(`burst (\d+b(?:/\d+)?)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["burst"] = matches[1]
		}
		if matches := regexp.MustCompile(`lat ([\d\.]+\w?s)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["latency"] = matches[1]
		}
	case "sfq":
		// SFQ also has optional parts like perturb
		if matches := regexp.MustCompile(`limit (\d+p)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["limit"] = matches[1]
		}
		if matches := regexp.MustCompile(`quantum (\d+b)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["quantum"] = matches[1]
		}
		if matches := regexp.MustCompile(`divisor (\d+)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["divisor"] = matches[1]
		}
		// The 'perturb' parameter is optional.
		if matches := regexp.MustCompile(`perturb (\d+sec)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["perturb"] = matches[1]
		}
	case "fq_codel":
		if matches := regexp.MustCompile(`limit (\d+p)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["limit"] = matches[1]
		}
		if matches := regexp.MustCompile(`flows (\d+)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["flows"] = matches[1]
		}
		if matches := regexp.MustCompile(`quantum (\d+)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["quantum"] = matches[1]
		}
		if matches := regexp.MustCompile(`target ([\d\.]+\w?s)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["target"] = matches[1]
		}
		if matches := regexp.MustCompile(`interval ([\d\.]+\w?s)`).FindStringSubmatch(rawOutput); len(matches) > 1 {
			settings["interval"] = matches[1]
		}
	}

	return algorithm, settings, nil
}
