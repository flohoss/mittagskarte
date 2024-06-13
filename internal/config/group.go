package config

import (
	"encoding/json"
	"errors"
)

type Group string

const (
	Degerloch              Group = "Degerloch"
	Fasanenhof             Group = "Fasanenhof"
	Feuerbach              Group = "Feuerbach"
	Koengen                Group = "Köngen"
	LeinfeldenEchterdingen Group = "Leinfelden-Echterdingen"
	Nuertingen             Group = "Nürtingen"
)

var allGroups = []Group{Degerloch, Fasanenhof, Feuerbach, Koengen, LeinfeldenEchterdingen, Nuertingen}

func (g *Group) UnmarshalJSON(data []byte) error {
	var group string
	if err := json.Unmarshal(data, &group); err != nil {
		return err
	}

	for _, validGroup := range allGroups {
		if Group(group) == validGroup {
			*g = Group(group)
			return nil
		}
	}
	return errors.New("invalid group")
}

func (g Group) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(g))
}
