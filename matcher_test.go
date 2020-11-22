package main_test

import (
	"fmt"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/r3labs/diff/v2"
	"sigs.k8s.io/yaml"
)

var _ = Describe("ContainsYAML", func() {
	It("contains keys with same values", func() {
		Expect(`a: 1`).To(ContainsYAML(`a: 1`))
		Expect(`a: 1`).NotTo(ContainsYAML(`b: 2`))

		Expect(`{a: 1, b: 2}`).To(ContainsYAML(`a: 1`))
		Expect(`{a: 1, b: 2}`).NotTo(ContainsYAML(`c: 3`))
	})

	It("supports nested values", func() {
		Expect(`{a: {b: 2, c: 3}, d: 4}`).To(ContainsYAML(`{a: {b: 2}}`))
		Expect(`{a: {b: 2, c: 3}, d: 4}`).ToNot(ContainsYAML(`{a: {b: 1}}`))
	})
})

type containsYAMLMatcher struct {
	expected interface{}
	log      interface{}
}

func (c *containsYAMLMatcher) yamlUnmarshal(value interface{}) (interface{}, error) {
	var contents []byte

	switch v := value.(type) {
	case string:
		contents = []byte(v)
	case []byte:
		contents = v
	case fmt.Stringer:
		contents = []byte(v.String())
	default:
		return nil, fmt.Errorf("requires a string, stringer, or []byte.  Got actual:\n%s", format.Object(value, 1))
	}

	var payload interface{}

	err := yaml.Unmarshal(contents, &payload)
	if err != nil {
		return nil, fmt.Errorf("String '%s' should be valid YAML, but it is not.\nUnderlying error: %w", contents, err)
	}

	return payload, nil
}

func (c *containsYAMLMatcher) Match(actual interface{}) (success bool, err error) {
	actualYAML, err := c.yamlUnmarshal(actual)
	if err != nil {
		return false, fmt.Errorf("ContainsYAML failed with actual value: %w", err)
	}

	expectedYAML, err := c.yamlUnmarshal(c.expected)
	if err != nil {
		return false, fmt.Errorf("ContainsYAML failed with expected value: %w", err)
	}

	changelog, err := diff.Diff(expectedYAML, actualYAML)
	if err != nil {
		return false, fmt.Errorf("ContainsYAML failed with comparing actual and expected: %w", err)
	}

	for _, log := range changelog {
		if log.Type != diff.CREATE {
			c.log = log

			return false, nil
		}
	}

	return true, nil
}

func (c *containsYAMLMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("actual to contain exepect, but had trouble with %s", c.log)
}

func (c *containsYAMLMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("actual to not contain exepect, but had trouble with %s", c.log)
}

func ContainsYAML(expected interface{}) types.GomegaMatcher {
	return &containsYAMLMatcher{
		expected: expected,
	}
}
