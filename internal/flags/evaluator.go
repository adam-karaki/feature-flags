package flags

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strconv"
)

func Evaluate(flag *Flag, input EvaluationContext) EvaluationResult {
	if !flag.Enabled {
		return EvaluationResult{Enabled: false, Reason: "flag_disabled", Version: flag.Version}
	}

	for _, rule := range flag.Rules {
		if !matches(rule, input) {
			continue
		}
		if rule.Percentage <= 0 || rule.Percentage >= 100 {
			return EvaluationResult{Enabled: true, Reason: "rule_match", Version: flag.Version}
		}
		bucket := percentageBucket(flag.Name, input.SubjectID)
		if bucket < rule.Percentage {
			return EvaluationResult{Enabled: true, Reason: "percentage_match", Version: flag.Version}
		}
		return EvaluationResult{Enabled: false, Reason: "percentage_excluded", Version: flag.Version}
	}

	return EvaluationResult{Enabled: false, Reason: "no_rule_match", Version: flag.Version}
}

func matches(rule Rule, input EvaluationContext) bool {
	value, ok := input.Attributes[rule.Attribute]
	if !ok {
		return false
	}
	switch rule.Operator {
	case "equals":
		return value == rule.Value
	case "not_equals":
		return value != rule.Value
	case "contains":
		return contains(value, rule.Value)
	default:
		return false
	}
}

func contains(value, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func percentageBucket(flagName, subjectID string) int {
	h := sha256.Sum256([]byte(flagName + ":" + subjectID))
	n := binary.BigEndian.Uint64(h[:8])
	return int(n % 100)
}

func ValidateRule(rule Rule) error {
	if rule.Attribute == "" {
		return fmt.Errorf("attribute is required")
	}
	if rule.Operator != "equals" && rule.Operator != "not_equals" && rule.Operator != "contains" {
		return fmt.Errorf("unsupported operator %q", rule.Operator)
	}
	if rule.Percentage < 0 || rule.Percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}
	return nil
}

func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > 128 {
		return fmt.Errorf("name must be at most 128 characters")
	}
	return nil
}

func ParseVersion(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}
