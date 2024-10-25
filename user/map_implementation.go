package user

import (
	"errors"
)

type MapStore struct {
	data map[string]*User // maps are already passed by ref
}

func NewMapStore() *MapStore {
	return &MapStore{
		data: make(map[string]*User),
	}
}

func (m *MapStore) List() ([]*User, error) {
	usersSlice := make([]*User, 0)

	for _, value := range m.data {
		usersSlice = append(usersSlice, value)
	}

	return usersSlice, nil
}

func (m *MapStore) Get(username string) (*User, error) {
	user, ok := m.data[username]
	if !ok {
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (m *MapStore) Create(username string, password string) (*User, error) {
	if _, ok := m.data[username]; ok {
		return nil, errors.New("user does not exist")
	}

	user := User{
		Username: username,
		Password: password,
	}

	m.data[username] = &user

	return &user, nil
}
