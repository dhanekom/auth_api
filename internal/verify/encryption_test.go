package verify

import "testing"

func TestPasswordEncryption(t *testing.T) {
	password := "1234"
	encryptor := PasswordEncryptorBcrypt{}
	passwordHashBytes, err := encryptor.GenerateHashedPassword(password)
	if err != nil {
		t.Errorf("unexpected error while generating password hash: %s", err.Error())
	}

	err = encryptor.CompareHashAndPassword(passwordHashBytes, []byte(password))
	if err != nil {
		t.Errorf("unexpected error while comparing password: %s", err.Error())
	}
}
