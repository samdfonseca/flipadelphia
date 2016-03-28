package store

type KeyValuePair [][]byte

type Serializable interface {
	Serialize() []byte
}

type PersistenceStore interface {
	Get([]byte, []byte) (Serializable, error)
	GetScopeFeatures([]byte) (Serializable, error)
	GetScopeFeaturesFilterByValue([]byte, []byte) (Serializable, error)
	Set([]byte, []byte, []byte) (Serializable, error)
	GetScopes() (Serializable, error)
	GetScopesWithPrefix([]byte) (Serializable, error)
	GetScopesWithFeature([]byte) (Serializable, error)
	GetFeatures() (Serializable, error)
}
