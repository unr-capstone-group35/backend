package user

type MapStore struct {
	data map[string]*User // maps are already passed by ref
}

func NewMapStore() *MapStore {
	return &MapStore{
		data: make(map[string]*User),
	}
}

func (m *MapStore) Get(username string) (*User, bool) {
	user, ok := m.data[username]
	if !ok {
		return nil, false
	}

	return user, true
}

func (m *MapStore) Create(username string, password string) bool {
	if _, ok := m.data[username]; ok {
		return false
	}

	user := User{
		Username: username,
		Password: password,
	}

	m.data[username] = &user

	return true
}
