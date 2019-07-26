package keys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ethereum/go-ethereum/console"
)

// promptPassphrase prompts the user for a passphrase.  Set confirmation to true
// to require the user to confirm the passphrase.
func promptPassphrase(confirmation bool) (string, error) {
	passphrase, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		return "", fmt.Errorf("Failed to read passphrase: %v", err)
	}

	if confirmation {
		confirm, err := console.Stdin.PromptPassword("Repeat passphrase: ")
		if err != nil {
			return "", fmt.Errorf("Failed to read passphrase confirmation: %v", err)
		}
		if passphrase != confirm {
			return "", fmt.Errorf("Passphrases do not match")
		}
	}

	return passphrase, nil
}

// getPassphrase obtains a passphrase given by the user.  It first checks the
// --passfile command line flag and ultimately prompts the user for a
// passphrase.
func getPassphrase() (string, error) {
	// Look for the --passfile flag.
	if passwordFile != "" {
		content, err := ioutil.ReadFile(passwordFile)
		if err != nil {
			return "", fmt.Errorf("Failed to read passphrase file '%s': %v", passwordFile, err)
		}
		return strings.TrimRight(string(content), "\r\n"), nil
	}

	// Otherwise prompt the user for the passphrase.
	return promptPassphrase(false)
}

// mustPrintJSON prints the JSON encoding of the given object and
// exits the program with an error message when the marshaling fails.
func mustPrintJSON(jsonObject interface{}) error {
	str, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON object: %v", err)
	}
	fmt.Println(string(str))
	return nil
}
