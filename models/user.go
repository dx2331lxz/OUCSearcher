package models

import (
	"OUCSearcher/database"
	"database/sql"
	"log"
)

type User struct {
	ID   int
	Name string
	Age  int
}

// Insert inserts a new user into the database
func InsertUser(name string, age int) (sql.Result, error) {
	query := "INSERT INTO users (name, age) VALUES (?, ?)"
	result, err := database.DB.Exec(query, name, age)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetAllUsers retrieves all users from the database
func GetAllUsers() ([]User, error) {
	query := "SELECT id, name, age FROM users"
	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Failed to query users: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Name, &user.Age)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
