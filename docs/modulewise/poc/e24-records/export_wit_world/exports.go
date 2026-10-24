package export_wit_world

import "wit_component/wit_world"

// Problems reports each named ID that no cited source carries, so the round trip
// of a string, a list of records and a list of strings is visible in the result.
func Problems(sentence string, cited []wit_world.Source, ids []string) []string {
	have := map[string]bool{}
	for _, s := range cited {
		have[s.Id] = true
	}
	out := []string{}
	for _, id := range ids {
		if !have[id] {
			out = append(out, "cites "+id+", which is not a cited source")
		}
	}
	return append(out, "checked: "+sentence)
}
