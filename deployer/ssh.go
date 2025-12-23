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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// generateSSHKey generates a new Ed25519 SSH key pair and saves it to the specified path.
// It returns the public key bytes in authorized_keys format.
func generateSSHKey(privateKeyPath string) ([]byte, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key: %v", err)
	}

	// Marshal private key to OpenSSH format
	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(block)
	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return nil, fmt.Errorf("failed to write private key: %v", err)
	}

	// Generate public key in authorized_keys format
	sshPubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh public key: %v", err)
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(sshPubKey)
	if err := os.WriteFile(privateKeyPath+".pub", pubKeyBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write public key: %v", err)
	}

	return pubKeyBytes, nil
}
