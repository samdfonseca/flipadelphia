package store

import "fmt"

var testFeatures = []FlipadelphiaSetFeatureOptions{
	FlipadelphiaSetFeatureOptions{"scope1", "feature1", "on"},
	FlipadelphiaSetFeatureOptions{"scope1", "feature2", "on"},
	FlipadelphiaSetFeatureOptions{"scope1", "feature3", "off"},
	FlipadelphiaSetFeatureOptions{"scope1", "feature4", "on"},
	FlipadelphiaSetFeatureOptions{"scope1", "feature5", "0"},
	FlipadelphiaSetFeatureOptions{"scope1", "feature6", "ON"},
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
