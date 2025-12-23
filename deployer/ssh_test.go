/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deployer

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerateSSHKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ssh-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	privateKeyPath := filepath.Join(tmpDir, "id_ed25519")
	pubKeyBytes, err := generateSSHKey(privateKeyPath)
	if err != nil {
		t.Fatalf("generateSSHKey failed: %v", err)
	}

	// Verify private key file exists
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Errorf("private key file does not exist")
	}

	// Verify public key file exists
	if _, err := os.Stat(privateKeyPath + ".pub"); os.IsNotExist(err) {
		t.Errorf("public key file does not exist")
	}

	// Verify public key format
	_, _, _, _, err = ssh.ParseAuthorizedKey(pubKeyBytes)
	if err != nil {
		t.Errorf("failed to parse generated public key: %v", err)
	}

	// Verify private key format
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatalf("failed to read private key: %v", err)
	}
	_, err = ssh.ParseRawPrivateKey(privBytes)
	if err != nil {
		t.Errorf("failed to parse generated private key: %v", err)
	}
}
