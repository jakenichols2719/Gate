package database

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

/* From config.go: DBConfig
Path string
*/

type UserCred struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UserPerm map[string]bool

type User struct {
	Credentials UserCred `json:"credentials"`
	Permissions UserPerm `json:"permissions"`
}

// UserEntry: User representation of the database
type UserEntry struct {
	ID           uint   `gorm:"autoIncrement,primaryKey"`
	Email        string `gorm:"email"`
	Username     string `gorm:"username"`
	PasswordHash string `gorm:"password"`
	Salt         string `gorm:"salt"`
	HashFunc     string `gorm:"hashfunc"`
	Permissions  string `gorm:"permissions"`
}

// Register a user with credentials and permissions
// Returns error or nil
func RegisterUser(email, username, password string, permissions UserPerm) error {
	if user, _ := findUserByEmail(email); user.Email != "" {
		return errors.New("Email is already in use.")
	}
	if user, _ := findUserByUsername(username); user.Username != "" {
		return errors.New("Username is already in use.")
	}
	perm, err := json.Marshal(permissions)
	salt, err := genSalt()
	if err != nil {
		return err
	}
	pwdHash, err := slowHash([]byte(password), salt, "sha512")
	if err != nil {
		return err
	}
	entry := &UserEntry{
		Email:        email,
		Username:     username,
		PasswordHash: pwdHash,
		Salt:         stringEncode(salt),
		HashFunc:     "sha512",
		Permissions:  string(perm),
	}
	addUser(entry)
	return nil
}

// Validate a user with credentials
// Returns success, user info, and error/nil
func ValidateUserCred(username, password string) (bool, *User, error) {
	// Find user. Fail out if non-eistent
	user, err := findUserByUsername(username)
	if err != nil {
		return false, nil, err
	}
	if user.Username == "" {
		return false, nil, fmt.Errorf("User %s not found", username)
	}
	// Check hashed password against input password.
	salt, _ := stringDecode(user.Salt)
	pwdHash, err := slowHash([]byte(password), salt, user.HashFunc)
	if err != nil {
		return false, nil, err
	}
	valid := username == user.Username && pwdHash == user.PasswordHash
	if valid {
		outUser := &User{
			Credentials: UserCred{
				user.Email,
				user.Username,
			},
			Permissions: make(UserPerm),
		}
		json.Unmarshal([]byte(user.Permissions), &outUser.Permissions)
		return true, outUser, nil
	} else {
		return false, &User{}, fmt.Errorf("User validation failed")
	}
}

// Change user password
// Returns success and error/nil
func ChangeUserPassword(username, password string, newPassword string) error {
	valid, _, err := ValidateUserCred(username, password)
	if !valid {
		if err != nil {
			return err
		} else {
			return fmt.Errorf("User validation failed")
		}
	}
	user, err := findUserByUsername(username)
	if err != nil {
		return err
	}
	// If the user didn't exist, validation would have failed
	salt, _ := genSalt()
	pwdHash, err := slowHash([]byte(newPassword), salt, user.HashFunc)
	if err != nil {
		return err
	}
	user.PasswordHash = pwdHash
	user.Salt = stringEncode(salt)
	updateUser(&user)
	return nil
}
