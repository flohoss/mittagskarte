package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"googlemaps.github.io/maps"
)

func TestGetMapInformation(t *testing.T) {
	info := GetMapInformation("AIzaSyCPVvOqLQTWu3_2Chr_9eqfZefxolRrUc8", []MapRequest{{
		Identifier: "test",
		Address:    "ENBW City, Fasanenhof",
	}, {
		Identifier: "bad",
		Address:    "",
	}})

	assert.IsType(t, "", info["test"].getLeg().HumanReadable)
	assert.NotEmpty(t, info["test"].getLeg().HumanReadable)

	assert.Empty(t, info["bad"].getLeg().HumanReadable)
	assert.IsType(t, &maps.Leg{}, info["bad"].getLeg())
}
