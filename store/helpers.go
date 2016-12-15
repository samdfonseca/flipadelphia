package store

import "fmt"

var testFeatures = []FlipadelphiaSetFeatureOptions{
	{"scope1", "feature1", "on"},
	{"scope1", "feature2", "on"},
	{"scope1", "feature3", "off"},
	{"scope1", "feature4", "on"},
	{"scope1", "feature5", "0"},
	{"scope1", "feature6", "ON"},
}

func ValidActivatedFeature(scope, key []byte) (Serializable, error) {
	return FlipadelphiaFeature{
		Name:  fmt.Sprintf("%s", key),
		Value: "on",
		Data:  "true",
	}, nil
}

func ValidUnactivatedFeature(scope, key []byte) (Serializable, error) {
	return FlipadelphiaFeature{
		Name:  fmt.Sprintf("%s", key),
		Value: "",
		Data:  "false",
	}, nil
}
