// Copyright 2018 OVO Technology
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypt

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

func init() {
	//defaultOptions := Defaults{}
	//parser := flags.NewParser(&defaultOptions, flags.Default)
	Parser.AddCommand("encrypt",
		"Encrypts your data, returning everything required for future decryption",
		"Creates a new DEK, encrypts data with DEK, encrypts the DEK using KMS, "+
			"spits out encrypted data + encrypted DEK.",
		&encryptCommand)
}

//EncryptCommand type
type EncryptCommand struct {
	DisableValidation bool   `short:"d" long:"disableValidation" description:"Disable validation of ciphertext"`
	Filepath          string `short:"f" long:"filepath" description:"Path of file to encrypt" default:"./plain.txt"`
	SingleLine        bool   `short:"s" long:"singleLine" description:"Disable use of newline chars in ciphertext"`
}

var encryptCommand EncryptCommand

//Execute executes the EncryptCommand
func (x *EncryptCommand) Execute(args []string) (err error) {
	fmt.Println("Encrypting...")
	dat, err := ioutil.ReadFile(x.Filepath)
	check(err)
	err = CipherText(dat, x.Filepath, x.SingleLine, x.DisableValidation)
	check(secureDelete(x.Filepath, false))
	return err
}

//insertNewLines inserts a newline char at specific intervals
func insertNewLines(cipherTexts []byte) (newLineText []byte) {
	interval := 40
	for i, char := range cipherTexts {
		if i > 0 && (i%interval == 0) {
			newLineText = append(newLineText, []byte("\n")...)
		}
		newLineText = append(newLineText, char)
	}
	return
}

//CipherText creates a ciphertext encrypted from a slice of bytes
//(the plaintext), and writes to File and Console.
func CipherText(plaintext []byte, filepath string, singleLine, disableValidation bool) (err error) {
	outputFilepath := "./cipher.txt"
	fileMode := os.FileMode.Perm(0644)
	cipherBytes := CipherBytes(plaintext, singleLine, disableValidation)
	fmt.Println("-----BEGIN (ENCRYPTED DATA + DEK) STRING-----")
	fmt.Printf("%s\n", cipherBytes)
	fmt.Println("-----END (ENCRYPTED DATA + DEK) STRING-----")
	ioutil.WriteFile(outputFilepath, cipherBytes, fileMode)
	fmt.Printf("Encryption successful, ciphertext available at %s\n",
		outputFilepath)
	return
}

//CipherBytes uses 'defaultOptions' go-flags to encrypt plaintext bytes and
//return ciphertext bytes
func CipherBytes(plaintext []byte, singleLine, disableValidation bool) (cipherBytes []byte) {

	kmsProvider, err := getKmsProvider(defaultOptions.KMSProvider)
	check(err)
	return CipherBytesFromPrimitives(plaintext, singleLine, disableValidation, defaultOptions.ProjectID,
		defaultOptions.LocationID, defaultOptions.KeyRingID,
		defaultOptions.CryptoKeyID, defaultOptions.KeyName, kmsProvider)
}

//CipherBytesFromPrimitives encrypts plaintext bytes and returns ciphertext bytes
func CipherBytesFromPrimitives(plaintext []byte, singleLine,
	disableValidation bool,
	projectID, locationID, keyRingID, cryptoKeyID, keyName string,
	kmsProvider KmsProvider) (cipherBytes []byte) {

	dek := randByteSlice(dekLength)
	nonce := randByteSlice(nonceLength)
	encrypt := true
	encryptedDek, err := kmsProvider.crypto(dek, projectID, locationID, keyRingID,
		cryptoKeyID, keyName, encrypt)
	check(err)
	cipherBytes = []byte(base64.StdEncoding.EncodeToString(append(
		append(cipherText(plaintext, cipherblock(dek), nonce, encrypt),
			nonce...),
		encryptedDek...)))
	if !singleLine {
		cipherBytes = insertNewLines(cipherBytes)
	}
	if !disableValidation {
		//validate the ciphertext
		fmt.Println("Validating ciphertext")
		cipherString, err := base64.StdEncoding.DecodeString(string(cipherBytes))
		check(err)
		_, err = PlainTextFromPrimitives(cipherString, projectID,
			locationID, keyRingID, cryptoKeyID, keyName, kmsProvider)
		check(err)
	}
	return
}
