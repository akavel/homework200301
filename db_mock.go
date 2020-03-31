package main

import (
	"fmt"
	"sync"
	"time"
)

type MockDB struct {
	users []*User
	m     sync.Mutex
}

func (db *MockDB) ListUsers() ([]*User, error) {
	db.m.Lock()
	defer db.m.Unlock()

	var snapshot []*User
	for _, u := range db.users {
		// NOTE: copying by value, to avoid race conditions when fields are modified
		clone := *u
		snapshot = append(snapshot, &clone)
	}
	return snapshot, nil
}

func (db *MockDB) GetUser(email string) (*User, error) {
	db.m.Lock()
	defer db.m.Unlock()

	u := db.findActive(email)
	if u == nil {
		return nil, nil
	}
	// NOTE: copying by value, to avoid race conditions when fields are modified
	clone := *u
	return &clone, nil
}

func (db *MockDB) CreateUser(u *User) error {
	db.m.Lock()
	defer db.m.Unlock()

	// Make sure user with the same email doesn't already exist
	conflict := db.findActive(u.Email)
	if conflict != nil {
		// FIXME: distinct error type for conflict
		return fmt.Errorf("user with the same .email already exists: %s", u.Email)
	}
	clone := *u
	db.users = append(db.users, &clone)
	return nil
}

func (db *MockDB) DeleteUser(email string) error {
	db.m.Lock()
	defer db.m.Unlock()

	u := db.findActive(email)
	if u == nil {
		// FIXME: distinct error type for 'not found'
		return fmt.Errorf("user not found: %s", email)
	}

	u.Deleted = newTime(time.Now())
	return nil
}

func (db *MockDB) Close() error { return nil }

func (db *MockDB) findActive(email string) *User {
	for _, u := range db.users {
		if u.Deleted == nil && u.Email == email {
			return u
		}
	}
	return nil
}

func NewMockDB() *MockDB {
	return &MockDB{
		users: []*User{
			{
				Name: "John", Surname: "Smith",
				Email:    "john@smith.name",
				Password: "iAmJohnny",
				// FIXME: provide location
				Birthday:   time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
				Address:    "some city, some street 1",
				Phone:      nil,
				Technology: "php",
				Deleted:    nil,
			},
			{
				Name: "Anne J. Marie", Surname: "von Flick",
				Email:    "Anne-Marie@vonflick.de",
				Password: `foo123!"%`,
				// FIXME: provide location
				Birthday:   time.Date(2000, 12, 12, 0, 0, 0, 0, time.UTC),
				Address:    "Anneville",
				Phone:      newString("+48 123-456-789"),
				Technology: "java",
				Deleted:    nil,
			},
			{
				Name:     "Robert'); DROP TABLE Users;--",
				Surname:  "Tables",                // TODO: fun surname (Unicode? Zalgo?)
				Email:    "bobby.tables@xkcd.com", // TODO: fun email
				Password: "TODO: harder",          // TODO: fun password
				// FIXME: provide location
				Birthday:   time.Date(2007, 10, 7, 0, 0, 0, 0, time.UTC), // TODO: fun date (Feb?)
				Address:    "Wherever",                                   // TODO: fun address (Unicode? Zalgo?)
				Phone:      nil,                                          // TODO: fun phone
				Technology: "js",
				Deleted:    nil,
			},
			{
				Name: "Dorothy", Surname: "Deleted de 1",
				Email:    "bobby.tables@xkcd.com",
				Password: "",
				// FIXME: provide location
				Birthday:   time.Date(2000, 12, 12, 0, 0, 0, 0, time.UTC),
				Address:    "Dorothyville 1",
				Phone:      newString("+48 1"),
				Technology: "go",
				// FIXME: provide location?
				Deleted: newTime(time.Date(2020, 3, 28, 12, 21, 1, 0, time.UTC)),
			},
			{
				Name: "Dorothy", Surname: "Deleted de 2",
				Email:    "bobby.tables@xkcd.com",
				Password: "",
				// FIXME: provide location
				Birthday:   time.Date(2000, 12, 12, 0, 0, 0, 0, time.UTC),
				Address:    "Dorothyville 2",
				Phone:      newString("+48 2"),
				Technology: "go",
				// FIXME: provide location?
				Deleted: newTime(time.Date(2020, 3, 28, 12, 21, 2, 0, time.UTC)),
			},
		},
	}
}

func newString(v string) *string     { return &v }
func newTime(v time.Time) *time.Time { return &v }
