package provider

import (
	"strings"
	"testing"
)

// The exact v7 definition our acceptance test submits.
const integrationSubmittedV7 = `
definitionVersion: 7
name: Acceptance Test
description: Acceptance Test Integration
requiredConfigVars: []
flows:
  - name: Flow 1
    isSynchronous: false
    steps:
      - name: Integration Trigger
        isTrigger: true
        action:
          component:
            key: webhook-triggers
            version: LATEST
            isPublic: true
          key: webhook
        inputs: {}`

// The normalized definition the API returns on read-back (captured via the live
// GraphQL API): version LATEST is resolved to 3, defaults are injected, and the
// empty requiredConfigVars is dropped.
const integrationCanonicalV7 = `"category": ""
"configPages":
- "elements": []
  "name": "Configuration"
"defaultInstanceProfile": "Default Instance Profile"
"definitionVersion": !!int "7"
"description": "Acceptance Test Integration"
"documentation": ""
"endpointType": "flow_specific"
"flows":
- "endpointSecurityType": "customer_optional"
  "isAgentFlow": !!bool "false"
  "isSynchronous": !!bool "false"
  "name": "Flow 1"
  "steps":
  - "action":
      "component":
        "isPublic": !!bool "true"
        "key": "webhook-triggers"
        "version": !!int "3"
      "key": "webhook"
    "description": ""
    "inputs": {}
    "isTrigger": !!bool "true"
    "name": "Integration Trigger"
"name": "Acceptance Test"`

func TestSuppressDiffIntegrationDefinition(t *testing.T) {
	cases := []struct {
		name       string
		old        string // normalized definition in state
		new        string // definition from config
		suppressed bool
	}{
		{
			name:       "submitted is a subset of the normalized read-back",
			old:        integrationCanonicalV7,
			new:        integrationSubmittedV7,
			suppressed: true,
		},
		{
			name:       "identical definitions",
			old:        integrationCanonicalV7,
			new:        integrationCanonicalV7,
			suppressed: true,
		},
		{
			name:       "changed action key is a real diff",
			old:        integrationCanonicalV7,
			new:        strings.Replace(integrationSubmittedV7, "key: webhook\n", "key: changed\n", 1),
			suppressed: false,
		},
		{
			name:       "changed name is a real diff",
			old:        integrationCanonicalV7,
			new:        strings.Replace(integrationSubmittedV7, "name: Acceptance Test\n", "name: Renamed\n", 1),
			suppressed: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := suppressDiffIntegrationDefinition("definition", tc.old, tc.new, nil); got != tc.suppressed {
				t.Errorf("suppress = %v, want %v", got, tc.suppressed)
			}
		})
	}
}

func TestDefinitionsEquivalent(t *testing.T) {
	cases := []struct {
		name      string
		submitted string
		canonical string
		want      bool
	}{
		{"version LATEST matches a resolved version",
			"action:\n  component:\n    version: LATEST",
			"action:\n  component:\n    version: 3",
			true},
		{"empty list matches an absent key",
			"name: X\nrequiredConfigVars: []",
			"name: X",
			true},
		{"empty string matches an absent key",
			"name: X\ndescription: \"\"",
			"name: X",
			true},
		{"injected server defaults are ignored",
			"name: X",
			"name: X\nendpointType: flow_specific\ndocumentation: \"\"",
			true},
		{"a non-empty value missing from canonical is a diff",
			"name: X\nfoo: bar",
			"name: X",
			false},
		{"a differing scalar is a diff",
			"version: 2",
			"version: 3",
			false},
		{"invalid yaml is never equivalent",
			"key: [unterminated",
			"name: X",
			false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := definitionsEquivalent(tc.submitted, tc.canonical); got != tc.want {
				t.Errorf("definitionsEquivalent = %v, want %v", got, tc.want)
			}
		})
	}
}
