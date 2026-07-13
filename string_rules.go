package main

import (
	"fmt"
	"regexp"
	"strings"
)

// StringMatchRuleConfig is the YAML-backed string match configuration. Regex values
// are compiled into StringMatchRule during runtime config preparation.
type StringMatchRuleConfig struct {
	Null        bool     `yaml:"empty"`
	Contains    string   `yaml:"contains"`
	ContainsAny []string `yaml:"contains_any"`
	ContainsAll []string `yaml:"contains_all"`
	Prefix      string   `yaml:"prefix"`
	Suffix      string   `yaml:"suffix"`
	RegexMatch  string   `yaml:"regex"`
}

// Returns true if StringMatchRuleConfig has any conditions
func (smr StringMatchRuleConfig) hasConditions() bool {
	return smr.Null ||
		smr.Contains != "" ||
		len(smr.ContainsAny) > 0 ||
		len(smr.ContainsAll) > 0 ||
		smr.Prefix != "" ||
		smr.Suffix != "" ||
		smr.RegexMatch != ""
}

func (smr StringMatchRuleConfig) compile() (StringMatchRule, error) {
	if err := validateStringList("contains_any", smr.ContainsAny); err != nil {
		return StringMatchRule{}, err
	}
	if err := validateStringList("contains_all", smr.ContainsAll); err != nil {
		return StringMatchRule{}, err
	}

	rule := StringMatchRule{
		Null:        smr.Null,
		Contains:    smr.Contains,
		ContainsAny: append([]string(nil), smr.ContainsAny...),
		ContainsAll: append([]string(nil), smr.ContainsAll...),
		Prefix:      smr.Prefix,
		Suffix:      smr.Suffix,
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

func validateStringList(name string, values []string) error {
	for i, value := range values {
		if value == "" {
			return fmt.Errorf("%s[%d] must not be empty", name, i)
		}
	}

	return nil
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
	Null        bool
	Contains    string
	ContainsAny []string
	ContainsAll []string
	Prefix      string
	Suffix      string
	Regex       *regexp.Regexp
}

func (rule StringMatchRule) hasConditions() bool {
	return rule.Null ||
		rule.Contains != "" ||
		len(rule.ContainsAny) > 0 ||
		len(rule.ContainsAll) > 0 ||
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
	// check contains_any if set
	if len(rule.ContainsAny) > 0 {
		if data == "" {
			return false
		}
		containsAny := false
		for _, value := range rule.ContainsAny {
			if strings.Contains(data, value) {
				containsAny = true
				break
			}
		}
		if !containsAny {
			return false
		}
	}
	// check contains_all if set
	if len(rule.ContainsAll) > 0 {
		if data == "" {
			return false
		}
		for _, value := range rule.ContainsAll {
			if !strings.Contains(data, value) {
				return false
			}
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
	Replace     string          `yaml:"replace"`
	Remove      bool            `yaml:"remove"`
	TrimPrefix  string          `yaml:"trim_prefix"`
	TrimSuffix  string          `yaml:"trim_suffix"`
	ReplaceText ReplaceTextRule `yaml:"replace_text"`
	Prefix      string          `yaml:"prefix"`
	Suffix      string          `yaml:"suffix"`
}

type ReplaceTextRule struct {
	Old string `yaml:"old"`
	New string `yaml:"new"`
	All bool   `yaml:"all"`
}

func (str StringTransformRule) hasActions() bool {
	return str.Replace != "" ||
		str.Remove ||
		str.TrimPrefix != "" ||
		str.TrimSuffix != "" ||
		str.ReplaceText.hasActions() ||
		str.Prefix != "" ||
		str.Suffix != ""
}

func (str StringTransformRule) validate() error {
	if str.ReplaceText.hasActions() && str.ReplaceText.Old == "" {
		return fmt.Errorf("replace_text.old must not be empty")
	}

	return nil
}

func (rule ReplaceTextRule) hasActions() bool {
	return rule.Old != "" || rule.New != "" || rule.All
}

func applyStringTransform(value string, rule StringTransformRule) string {
	if rule.Remove {
		return ""
	}
	if rule.Replace != "" {
		return rule.Replace
	}
	if rule.TrimPrefix != "" {
		value = strings.TrimPrefix(value, rule.TrimPrefix)
	}
	if rule.TrimSuffix != "" {
		value = strings.TrimSuffix(value, rule.TrimSuffix)
	}
	if rule.ReplaceText.Old != "" {
		n := 1
		if rule.ReplaceText.All {
			n = -1
		}
		value = strings.Replace(value, rule.ReplaceText.Old, rule.ReplaceText.New, n)
	}
	if rule.Prefix != "" {
		value = rule.Prefix + value
	}
	if rule.Suffix != "" {
		value += rule.Suffix
	}

	return value
}
