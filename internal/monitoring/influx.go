package monitoring

import "strings"

type point struct {
	table  string
	tags   map[string]string
	values map[string]string
}

func (p point) toRecord() string {
	return p.table + "," + transformMap(p.tags) + " " + transformMap(p.values)
}

func transformMap(m map[string]string) (str string) {
	for key, val := range m {
		str = str + key + "=" + val + ","
	}
	return strings.TrimSuffix(str, ",")
}
