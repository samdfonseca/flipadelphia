package store

type MockPersistenceStore struct {
	OnGet                           func([]byte, []byte) (Serializable, error)
	OnGetScopeFeatures              func([]byte) (Serializable, error)
	OnGetScopeFeaturesFilterByValue func([]byte, []byte) (Serializable, error)
	OnSet                           func([]byte, []byte, []byte) (Serializable, error)
	OnGetScopes                     func() (Serializable, error)
	OnGetScopesWithPrefix           func([]byte) (Serializable, error)
	OnGetScopesWithFeature          func([]byte) (Serializable, error)
	OnGetScopesPaginated            func(int, int) (Serializable, error)
	OnGetFeatures                   func() (Serializable, error)
	OnGetScopeFeaturesFull          func([]byte) (Serializable, error)
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

func (mStore MockPersistenceStore) GetScopes() (Serializable, error) {
	return mStore.OnGetScopes()
}

func (mStore MockPersistenceStore) GetScopesWithPrefix(prefix []byte) (Serializable, error) {
	return mStore.OnGetScopesWithPrefix(prefix)
}

func (mStore MockPersistenceStore) GetScopesWithFeature(feature []byte) (Serializable, error) {
	return mStore.OnGetScopeFeatures(feature)
}

func (mStore MockPersistenceStore) GetScopesPaginated(offset, count int) (Serializable, error) {
	return mStore.OnGetScopesPaginated(offset, count)
}

func (mStore MockPersistenceStore) GetFeatures() (Serializable, error) {
	return mStore.OnGetFeatures()
}

func (mStore MockPersistenceStore) GetScopeFeaturesFull(scope []byte) (Serializable, error) {
	return mStore.OnGetScopeFeaturesFull(scope)
}

func (mStore MockPersistenceStore) Close() error {
	return nil
}
