// OrderedMap: a mapping value whose keys keep the order they were given in.

package yamler

import "gopkg.in/yaml.v3"

// OrderedMap is a mapping whose keys keep the order they are listed in.
// Pass one to Set (or nest it inside a slice or another OrderedMap) when the
// key order of a value you create matters; a plain map[string]interface{} is
// written with its keys sorted, because Go maps have no order of their own.
//
//	doc.Set("database", yamler.OrderedMap{
//	    Keys:   []string{"host", "port", "name"},
//	    Values: map[string]interface{}{"host": "localhost", "port": 5432, "name": "app"},
//	})
//
// Keys listed in Keys but missing from Values are written with a null value;
// keys present in Values but not listed in Keys are not written at all.
type OrderedMap struct {
	Keys   []string
	Values map[string]interface{}
}

// Node builds the YAML mapping node for the map, in Keys order.
func (om OrderedMap) Node() (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, key := range om.Keys {
		valueNode, err := interfaceToNode(om.Values[key])
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, createScalarNode("!!str", key), valueNode)
	}
	return node, nil
}

// MarshalYAML implements yaml.Marshaler so that an OrderedMap keeps its key
// order when it is encoded by gopkg.in/yaml.v3 directly.
func (om OrderedMap) MarshalYAML() (interface{}, error) {
	return om.Node()
}
