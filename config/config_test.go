package config

import (
	"testing"

	"github.com/dankraw/ssh-aliases/compiler"
	"github.com/stretchr/testify/assert"
)

func TestShouldMapToHostConfigInputs(t *testing.T) {
	t.Parallel()

	// given
	config := RawConfigContext{
		Hosts: []Host{{
			Name:           "service-a",
			Hostname:       "service-a[1..5].example.com",
			Alias:          "a%1",
			RawConfigOrRef: "service-a",
		}, {
			Name:     "service-b",
			Hostname: "service-b[1..2].example.com",
			Alias:    "b%1",
			RawConfigOrRef: []map[string]interface{}{{
				"identity_file": "b_id_rsa.pub",
			}, {
				"port": 22,
			}},
		}}, RawConfigs: RawConfigs{
			"service-a": RawConfig{{
				"identity_file": "a_id_rsa.pub",
				"port":          22,
			}}},
	}

	// when
	inputs, err := config.ToExpandingHostConfigs()

	// then
	assert.NoError(t, err)
	assert.Equal(t, []compiler.ExpandingHostConfig{{
		AliasName:       "service-a",
		HostnamePattern: "service-a[1..5].example.com",
		AliasTemplate:   "a%1",
		Config: compiler.ConfigProperties{
			{"IdentityFile", "a_id_rsa.pub"},
			{"Port", 22},
		}}, {
		AliasName:       "service-b",
		HostnamePattern: "service-b[1..2].example.com",
		AliasTemplate:   "b%1",
		Config: compiler.ConfigProperties{
			{"IdentityFile", "b_id_rsa.pub"},
			{"Port", 22},
		}},
	}, inputs)
}

func TestShouldReturnErrorOnDuplicateKey(t *testing.T) {
	t.Parallel()

	// given
	config := RawConfigContext{
		Hosts: []Host{{
			Name:     "service-b",
			Hostname: "service-b[1..2].example.com",
			Alias:    "b%1",
			RawConfigOrRef: []map[string]interface{}{{
				"identity_file": "b_id_rsa.pub",
			}, {
				"identity_file": "c_id_rsa.pub",
			}},
		}}}

	// when
	inputs, err := config.ToExpandingHostConfigs()

	// then
	assert.Error(t, err)
	assert.Nil(t, inputs)
	assert.Equal(t, "Duplicate config property `identity_file` for host `service-b`", err.Error())
}

func TestShouldReturnErrorOnDuplicateKeyInRawConfigs(t *testing.T) {
	t.Parallel()

	// given
	config := RawConfigContext{
		RawConfigs: RawConfigs{
			"service-a": RawConfig{
				{"identity_file": "abc"},
				{"identity_file": "abc"},
			},
		},
	}

	// when
	inputs, err := config.ToExpandingHostConfigs()

	// then
	assert.Error(t, err)
	assert.Nil(t, inputs)
	assert.Equal(t, "Duplicate config entry `identity_file` in host `service-a`", err.Error())
}

func TestShouldReturnErrorOnNotFoundSSHConfig(t *testing.T) {
	t.Parallel()

	// given
	config := RawConfigContext{
		Hosts: []Host{{
			Name:           "service-a",
			Hostname:       "service-a[1..5].example.com",
			Alias:          "a%1",
			RawConfigOrRef: "this-does-not-exists",
		}},
	}

	// when
	results, err := config.ToExpandingHostConfigs()

	// then
	assert.Nil(t, results)
	assert.Error(t, err)
	assert.Equal(t, "No config `this-does-not-exists` found (used by host `service-a`)", err.Error())
}

func TestShouldMergeWithOtherConfig(t *testing.T) {
	t.Parallel()

	// given
	config := RawConfigContext{
		Hosts: []Host{{
			Name: "project1",
		}},
		RawConfigs: RawConfigs{
			"project1-config": RawConfig{{
				"identity_file": "a_id_rsa.pub",
			}},
		},
	}

	// when
	err := config.Merge(RawConfigContext{
		Hosts: []Host{{
			Name: "project2",
		}},
		RawConfigs: RawConfigs{
			"project2-config": RawConfig{{
				"port": 22,
			}},
		},
	})

	// then
	assert.NoError(t, err)
	assert.Equal(t, RawConfigContext{
		Hosts: []Host{{
			Name: "project1",
		}, {
			Name: "project2",
		}},
		RawConfigs: RawConfigs{
			"project1-config": RawConfig{{
				"identity_file": "a_id_rsa.pub",
			}},
			"project2-config": RawConfig{{
				"port": 22,
			}},
		},
	}, config)
}

func TestShouldReturnErrorOnDuplicateSSHConfigWhenMerging(t *testing.T) {
	t.Parallel()

	// given
	config := RawConfigContext{
		RawConfigs: RawConfigs{
			"service-a": RawConfig{{
				"identity_file": "a_id_rsa.pub",
			}},
		},
	}
	config2 := RawConfigContext{
		RawConfigs: RawConfigs{
			"service-a": RawConfig{{
				"port": 22,
			}},
		},
	}

	// when
	err := config.Merge(config2)

	// then
	assert.Error(t, err)
	assert.Equal(t, "Duplicate config `service-a`", err.Error())
}

func TestShouldReturnErrorOnDuplicateAlias(t *testing.T) {
	t.Parallel()

	// given
	config := RawConfigContext{
		Hosts: []Host{{
			Name: "service-a",
		}, {
			Name: "service-a",
		}},
	}

	// when
	results, err := config.ToExpandingHostConfigs()

	// then
	assert.Nil(t, results)
	assert.Error(t, err)
}

func TestShouldSortHostConfigAndSanitizeKeywords(t *testing.T) {
	t.Parallel()

	// given
	config := Config{
		"b": "something",
		"c": "abc",
		"d": 0,
		"a": 123,
	}

	// when
	entries := config.toSortedProperties()

	// then
	assert.Equal(t, compiler.ConfigProperties{
		{"A", 123},
		{"B", "something"},
		{"C", "abc"},
		{"D", 0},
	}, entries)
}
