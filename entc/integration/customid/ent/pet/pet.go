// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Code generated by entc, DO NOT EDIT.

package pet

const (
	// Label holds the string label denoting the pet type in the database.
	Label = "pet"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// EdgeCars holds the string denoting the cars edge name in mutations.
	EdgeCars = "cars"
	// EdgeFriends holds the string denoting the friends edge name in mutations.
	EdgeFriends = "friends"
	// EdgeBestFriend holds the string denoting the best_friend edge name in mutations.
	EdgeBestFriend = "best_friend"
	// UserFieldID holds the string denoting the ID field of the User.
	UserFieldID = "oid"
	// Table holds the table name of the pet in the database.
	Table = "pets"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "pets"
	// OwnerInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	OwnerInverseTable = "users"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "user_pets"
	// CarsTable is the table that holds the cars relation/edge.
	CarsTable = "cars"
	// CarsInverseTable is the table name for the Car entity.
	// It exists in this package in order to avoid circular dependency with the "car" package.
	CarsInverseTable = "cars"
	// CarsColumn is the table column denoting the cars relation/edge.
	CarsColumn = "pet_cars"
	// FriendsTable is the table that holds the friends relation/edge. The primary key declared below.
	FriendsTable = "pet_friends"
	// BestFriendTable is the table that holds the best_friend relation/edge.
	BestFriendTable = "pets"
	// BestFriendColumn is the table column denoting the best_friend relation/edge.
	BestFriendColumn = "pet_best_friend"
)

// Columns holds all SQL columns for pet fields.
var Columns = []string{
	FieldID,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "pets"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"pet_best_friend",
	"user_pets",
}

var (
	// FriendsPrimaryKey and FriendsColumn2 are the table columns denoting the
	// primary key for the friends relation (M2M).
	FriendsPrimaryKey = []string{"pet_id", "friend_id"}
)

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
	// IDValidator is a validator for the "id" field. It is called by the builders before save.
	IDValidator func(string) error
)
