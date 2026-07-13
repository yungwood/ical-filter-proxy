package main

import (
	"fmt"
	"regexp"
	"strings"
)

// StringMatchRuleConfig is the YAML-backed string match configuration. Regex values
// are compiled into StringMatchRule during runtime config preparation.
type StringMatchRuleConfig struct {
	Null       bool   `yaml:"empty"`
	Contains   string `yaml:"contains"`
	Prefix     string `yaml:"prefix"`
	Suffix     string `yaml:"suffix"`
	RegexMatch string `yaml:"regex"`
}

// Returns true if StringMatchRuleConfig has any conditions
func (smr StringMatchRuleConfig) hasConditions() bool {
	return smr.Null ||
		smr.Contains != "" ||
		smr.Prefix != "" ||
		smr.Suffix != "" ||
		smr.RegexMatch != ""
}

func (smr StringMatchRuleConfig) compile() (StringMatchRule, error) {
	rule := StringMatchRule{
		Null:     smr.Null,
		Contains: smr.Contains,
		Prefix:   smr.Prefix,
		Suffix:   smr.Suffix,
	}

	if smr.RegexMatch == "" {
		return rule, nil
	}

	regex, err := regexp.Compile(smr.RegexMatch)
	if err != nil {
		return rule, fmt.Errorf("invalid regex %q: %w", smr.RegexMatch, err)
	}
	rule.Regex = regex

	return rule, nil
}

// Returns true if a given string (data) matches ALL StringMatchRuleConfig conditions
func (smr StringMatchRuleConfig) matchesString(data string) bool {
	rule, err := smr.compile()
	if err != nil {
		return false
	}

	return rule.matchesString(data)
}

// StringMatchRule is the runtime string matcher. Regex is compiled once
// so event processing can match without reparsing configuration.
type StringMatchRule struct {
	Null     bool
	Contains string
	Prefix   string
	Suffix   string
	Regex    *regexp.Regexp
}

func (rule StringMatchRule) hasConditions() bool {
	return rule.Null ||
		rule.Contains != "" ||
		rule.Prefix != "" ||
		rule.Suffix != "" ||
		rule.Regex != nil
}

func (rule StringMatchRule) matchesString(data string) bool {
	// check null if set and don't process further - this condition can only be met on its own
	if rule.Null {
		return data == ""
	}
	// check contains if set
	if rule.Contains != "" {
		if data == "" || !strings.Contains(data, rule.Contains) {
			return false
		}
	}
	// check prefix if set
	if rule.Prefix != "" {
		if data == "" || !strings.HasPrefix(data, rule.Prefix) {
			return false
		}
	}
	// check suffix if set
	if rule.Suffix != "" {
		if data == "" || !strings.HasSuffix(data, rule.Suffix) {
			return false
		}
	}
	// check regex match if set
	if rule.Regex != nil {
		match := rule.Regex.MatchString(data)
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
