package store

type MockPersistenceStore struct {
	OnGet                           func([]byte, []byte) (Serializable, error)
	OnGetScopeFeatures              func([]byte) (Serializable, error)
	OnGetScopeFeaturesFilterByValue func([]byte, []byte) (Serializable, error)
	OnSet                           func([]byte, []byte, []byte) (Serializable, error)
}

func (mStore MockPersistenceStore) Get(scope, key []byte) (Serializable, error) {
	return mStore.OnGet(scope, key)
}

func (mStore MockPersistenceStore) GetScopeFeatures(scope []byte) (Serializable, error) {
	return mStore.OnGetScopeFeatures(scope)
}

func (mStore MockPersistenceStore) GetScopeFeaturesFilterByValue(scope, value []byte) (Serializable, error) {
	return mStore.OnGetScopeFeaturesFilterByValue(scope, value)
}

func (mStore MockPersistenceStore) Set(scope, key, value []byte) (Serializable, error) {
	return mStore.OnSet(scope, key, value)
}
