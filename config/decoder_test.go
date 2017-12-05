package config

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const fixtureDir = "./test-fixtures"

func TestShouldDecodeConfig(t *testing.T) {
	t.Parallel()

	// given
	data, _ := ioutil.ReadFile(filepath.Join(fixtureDir, "example.hcl"))

	// when
	config, _ := newDecoder().decode(data)

	// then
	assert.Equal(t, rawFileContext{
		Hosts: []host{{
			Name:           "service-a",
			Hostname:       "service-a[1..5].example.com",
			Alias:          "a{#1}",
			RawConfigOrRef: "service-a",
		}, {
			Name:     "service-b",
			Hostname: "service-b[1..2].example.com",
			Alias:    "b{#1}",
			RawConfigOrRef: []map[string]interface{}{{
				"identity_file": "b_id_rsa.pem",
			}, {
				"port": 22,
			}},
		}}, RawConfigs: rawConfigs{
			"service-a": rawConfig{{
				"identity_file": "a_id_rsa.pem",
				"port":          22,
			}},
		},
	}, config)

}
