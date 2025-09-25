package providers

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// We define specific, robust regexes for each algorithm we want to parse in detail.
// They are now less strict about formatting.
var (
	cakeRegex    = regexp.MustCompile(`qdisc cake.*?bandwidth (\d+[KMG]?bit)(.*)`)
	tbfRegex     = regexp.MustCompile(`qdisc tbf.*?rate (\d+[KMG]?bit) burst (\d+b) lat (\d+m?s)`)
	sfqRegex     = regexp.MustCompile(`qdisc sfq.*?limit (\d+p) quantum (\d+b) perturb (\d+sec)`)
	fqCodelRegex = regexp.MustCompile(`qdisc fq_codel.*?limit (\d+p) flows (\d+) quantum (\d+) target (\d+m?s) interval (\d+m?s)`)
)

// GetCurrentRule has been rewritten to be a sophisticated, multi-pass parser.
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

	settings := make(map[string]interface{})

	// --- Primary Detection: Check for the most specific qdiscs first ---
	if matches := cakeRegex.FindStringSubmatch(rawOutput); len(matches) > 1 {
		settings["bandwidth"] = matches[1]
		settings["options"] = strings.TrimSpace(matches[2])
		return "cake", settings, nil
	}
	if matches := tbfRegex.FindStringSubmatch(rawOutput); len(matches) > 1 {
		settings["rate"] = matches[1]
		settings["burst"] = matches[2]
		settings["latency"] = matches[3]
		return "tbf", settings, nil
	}
	if matches := sfqRegex.FindStringSubmatch(rawOutput); len(matches) > 1 {
		settings["limit"] = matches[1]
		settings["quantum"] = matches[2]
		settings["perturb"] = matches[3]
		return "sfq", settings, nil
	}

	// 【核心修正】 FQ Codel is often the most important one to detect correctly.
	if matches := fqCodelRegex.FindStringSubmatch(rawOutput); len(matches) > 1 {
		settings["limit"] = matches[1]
		settings["flows"] = matches[2]
		settings["quantum"] = matches[3]
		settings["target"] = matches[4]
		settings["interval"] = matches[5]
		if strings.Contains(rawOutput, "qdisc mq") {
			settings["parent"] = "mq (Multi-queue)"
		}
		return "fq_codel", settings, nil
	}

	// --- Fallback Detection: If no detailed match, check for simple presence ---
	if strings.Contains(rawOutput, "qdisc mq") {
		return "mq", map[string]interface{}{"note": "This is a parent qdisc, with child qdiscs (likely fq_codel) active below."}, nil
	}
	if strings.Contains(rawOutput, "qdisc pfifo_fast") {
		return "pfifo_fast", settings, nil
	}

	// If nothing matched, return the raw output for debugging.
	settings["raw_output"] = rawOutput
	return "unknown", settings, nil
}
