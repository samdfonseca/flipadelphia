package store

import "fmt"

var testFeatures = []FlipadelphiaSetFeatureOptions{
	FlipadelphiaSetFeatureOptions{[]byte("scope1"), []byte("feature1"), []byte("on")},
	FlipadelphiaSetFeatureOptions{[]byte("scope1"), []byte("feature2"), []byte("on")},
	FlipadelphiaSetFeatureOptions{[]byte("scope1"), []byte("feature3"), []byte("off")},
	FlipadelphiaSetFeatureOptions{[]byte("scope1"), []byte("feature4"), []byte("on")},
	FlipadelphiaSetFeatureOptions{[]byte("scope1"), []byte("feature5"), []byte("0")},
	FlipadelphiaSetFeatureOptions{[]byte("scope1"), []byte("feature6"), []byte("ON")},
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
