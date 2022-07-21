package database

import (
	"os"
	"testing"
)

func TestDBAccess(t *testing.T) {
	// Delete previous test database
	err := os.Remove("test.db")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// Open database

	if err := OpenDB("test.db"); err != nil {
		t.Error(err)
	}

	if valid, _, _ := ValidateUserCred("username", "password"); valid {
		t.Error("Validated user credentials that haven't been registered")
	}

	// Register a user
	if err := RegisterUser("user@email.com", "username", "password", nil); err != nil {
		t.Error(err)
	}

	// Validate those user credentials
	if _, _, err := ValidateUserCred("username", "password"); err != nil {
		t.Error(err)
	}

	// Validate again, but using the wrong password
	if valid, _, _ := ValidateUserCred("username", "wrongpassword"); valid {
		t.Error("Validated user credentials that are incorrect")
	}

	// Try re-registering (should fail)
	if err := RegisterUser("user@email.com", "username", "password", nil); err == nil {
		t.Error("Re-registration under same email/username succeeded when it should fail")
	}

	// Given that fails, let's change passwords. Wrong username?
	if err := ChangeUserPassword("wrongusername", "password", "newPassword"); err == nil {
		t.Error("Password change with incorrect username succeeded when it should fail")
	}

	// Wrong old password
	if err := ChangeUserPassword("username", "wrongpassword", "newPassword"); err == nil {
		t.Error("Password change with incorrect old password succeeded when it should fail")
	}

	// Correct everything
	if err := ChangeUserPassword("username", "password", "newPassword"); err != nil {
		t.Error(err)
	}

	// Re-validate with new password
	if _, _, err := ValidateUserCred("username", "newPassword"); err != nil {
		t.Error(err)
	}
}
