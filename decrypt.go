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

package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

func init() {
	parser.AddCommand("decrypt",
		"Decrypts encrypted text, returning the plaintext data",
		"Decrypts the encrypted DEK via KMS, decrypts the data with the DEK, "+
			"outputs to file",
		&decryptCommand)
}

//DecryptCommand type
type DecryptCommand struct {
	Filepath         string `short:"f" long:"filepath" description:"Path of file to get encrypted string from" default:"./cipher.txt"`
	RetainCipherText bool   `short:"r" long:"retainCipherText" description:"Retain ciphertext after decryption"`
	TargetFilepath   string `short:"t" long:"targetFilepath" description:"Path of file to write decrypted string to" default:"./plain.txt"`
	Validate         bool   `short:"v" long:"validate" description:"Validate decryption works"`
	WriteToStdout    bool   `short:"o" long:"stdout" description:"Writes decrypted plaintext to console"`
}

var decryptCommand DecryptCommand

//Execute executes the DecryptCommand
func (x *DecryptCommand) Execute(args []string) error {
	if !x.WriteToStdout {
		fmt.Println("Decrypting...")
	}
	plaintext, err := PlainText(x.Filepath)
	outputFilepath := x.TargetFilepath
	fileMode := os.FileMode.Perm(0644)
	if x.Validate {
		fmt.Println("Validation completed successfully")
		os.Exit(0)
	}
	if x.WriteToStdout {
		fmt.Printf("%s\n", plaintext)
	} else {
		err = ioutil.WriteFile(outputFilepath, plaintext, fileMode)
		check(err)
		fmt.Printf("Decryption successful, plaintext available at %s\n",
			outputFilepath)
	}
	if !x.RetainCipherText {
		check(secureDelete(x.Filepath, x.WriteToStdout))
	}
	return err
}

// PlainText returns a slice of bytes (the plaintext), decrypted from File
func PlainText(filepath string) (plaintext []byte, err error) {
	file, err := os.Open(filepath)
	check(err)
	defer file.Close()
	s := bufio.NewScanner(file)
	var buffer bytes.Buffer
	for s.Scan() {
		buffer.WriteString(s.Text())
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(buffer.String())
	check(err)
	dekLength := 113
	cipherLength := len(cipherBytes)
	encrypt := false
	encryptedDek := cipherBytes[cipherLength-dekLength : cipherLength]
	nonce := cipherBytes[cipherLength-(dekLength+nonceLength) : cipherLength-dekLength]
	decryptedDek := googleKMSCrypto(encryptedDek, defaultOptions.ProjectID,
		defaultOptions.LocationID, defaultOptions.KeyRingID,
		defaultOptions.CryptoKeyID, defaultOptions.KeyName, encrypt)
	plaintext = cipherText(cipherBytes[0:len(cipherBytes)-(dekLength+nonceLength)],
		cipherblock(decryptedDek), nonce, encrypt)
	return
}
