package store

type KeyValuePair [][]byte

type Serializable interface {
	Serialize() []byte
}

type PersistenceStore interface {
	Get([]byte, []byte) (Serializable, error)
	GetScopeFeatures([]byte) (Serializable, error)
	Set([]byte, []byte, []byte) (Serializable, error)
}
