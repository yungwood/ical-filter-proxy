package main

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// StringMatchRule defines match rules for VEvent properties with string values
type StringMatchRule struct {
	Null       bool   `yaml:"empty"`
	Contains   string `yaml:"contains"`
	Prefix     string `yaml:"prefix"`
	Suffix     string `yaml:"suffix"`
	RegexMatch string `yaml:"regex"`
}

// Returns true if StringMatchRule has any conditions
func (smr StringMatchRule) hasConditions() bool {
	return smr.Null ||
		smr.Contains != "" ||
		smr.Prefix != "" ||
		smr.Suffix != "" ||
		smr.RegexMatch != ""
}

func (smr StringMatchRule) validate() error {
	if smr.RegexMatch == "" {
		return nil
	}

	if _, err := regexp.Compile(smr.RegexMatch); err != nil {
		return fmt.Errorf("invalid regex %q: %w", smr.RegexMatch, err)
	}

	return nil
}

// Returns true if a given string (data) matches ALL StringMatchRule conditions
func (smr StringMatchRule) matchesString(data string) bool {
	// check null if set and don't process further - this condition can only be met on its own
	if smr.Null {
		return data == ""
	}
	// check contains if set
	if smr.Contains != "" {
		if data == "" || !strings.Contains(data, smr.Contains) {
			return false
		}
	}
	// check prefix if set
	if smr.Prefix != "" {
		if data == "" || !strings.HasPrefix(data, smr.Prefix) {
			return false
		}
	}
	// check suffix if set
	if smr.Suffix != "" {
		if data == "" || !strings.HasSuffix(data, smr.Suffix) {
			return false
		}
	}
	// check regex match if set
	if smr.RegexMatch != "" {
		re, err := regexp.Compile(smr.RegexMatch)
		if err != nil {
			slog.Warn("error processing regex rule", "value", smr.RegexMatch)
			return false // regex error is considered a failure to match
		}
		match := re.MatchString(data)
		if !match {
			return false // regex didn't match
		}
	}
	return true
}

// StringTransformRule defines changes for VEvent properties with string values
type StringTransformRule struct {
	Replace string `yaml:"replace"`
	Remove  bool   `yaml:"remove"`
	Prefix  string `yaml:"prefix"`
	Suffix  string `yaml:"suffix"`
}

func (str StringTransformRule) hasActions() bool {
	return str.Replace != "" ||
		str.Remove ||
		str.Prefix != "" ||
		str.Suffix != ""
}

func applyStringTransform(value string, rule StringTransformRule) string {
	if rule.Remove {
		return ""
	}
	if rule.Replace != "" {
		return rule.Replace
	}
	if rule.Prefix != "" {
		value = rule.Prefix + value
	}
	if rule.Suffix != "" {
		value += rule.Suffix
	}

	return value
}
