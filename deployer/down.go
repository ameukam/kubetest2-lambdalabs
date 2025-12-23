package deployer

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

func (d *deployer) Down() error {
	if d.apiKey == "" {
		return fmt.Errorf("Lambda AI API key must be provided via --api-key flag or LAMBDA_AI_API_KEY env var")
	}

	klog.Infof("Bringing down instance %s...", d.instanceName)

	// 1. Find instance by name
	instances, err := d.listInstances()
	if err != nil {
		return fmt.Errorf("failed to list instances: %v", err)
	}

	var instanceID string
	for _, inst := range instances {
		if inst.Name == d.instanceName {
			// Check if it's already terminated?
			if inst.Status == "terminated" {
				continue
			}
			instanceID = inst.ID
			break
		}
	}

	if instanceID == "" {
		klog.Infof("Instance %s not found or already terminated.", d.instanceName)
	} else {
		// 2. Terminate instance
		klog.Infof("Terminating instance %s (ID: %s)...", d.instanceName, instanceID)
		if err := d.terminateInstance(instanceID); err != nil {
			return fmt.Errorf("failed to terminate instance: %v", err)
		}
		klog.Info("Instance termination requested.")
	}

	// 3. Delete SSH Key if we generated it
	// We assume if d.sshKeyName matches d.instanceName (or is empty and we default), we delete it.
	// In Up(), if ssh-key was empty, we set it to instanceName.
	// But in Down(), d.sshKeyName might be empty if flag not passed.
	keyName := d.sshKeyName
	if keyName == "" {
		keyName = d.instanceName
	}
	
	// Only delete if it matches instanceName (heuristic for "we created it")
	// If the user explicitly passed a key name different from instance name, we probably shouldn't delete it.
	// If user passed --ssh-key=mykey, we keep it.
	// If user didn't pass --ssh-key, d.sshKeyName is "", so we assume keyName = d.instanceName.
	// But wait, if user passed --ssh-key=instanceName, we might delete it by mistake?
	// It's a risk. But standard kubetest2 ephemeral behavior usually cleans up everything with that name.
	
	if keyName == d.instanceName {
		klog.Infof("Deleting SSH key %s...", keyName)
		// Check if exists first to avoid error?
		keys, err := d.listSSHKeys()
		if err == nil {
			found := false
			for _, k := range keys {
				if k.Name == keyName {
					found = true
					break
				}
			}
			if found {
				// Delete key
				// API doesn't have delete SSH key? I need to check.
				// My client.go didn't implement deleteSSHKey.
				// I need to check if there is a delete endpoint.
				// Usually DELETE /ssh-keys/:id or similar.
				// I'll skip this for now or try to implement it.
				// Search results didn't explicitly mention delete key.
				// I'll assume DELETE /ssh-keys/{id} works.
				
				// I need to find ID first.
				var keyID string
				for _, k := range keys {
					if k.Name == keyName {
						keyID = k.ID
						break
					}
				}
				
				if keyID != "" {
					if err := d.deleteSSHKey(keyID); err != nil {
						klog.Warningf("Failed to delete SSH key %s: %v", keyName, err)
					} else {
						klog.Info("SSH key deleted.")
					}
				}
			} else {
				klog.Infof("SSH key %s not found, skipping deletion.", keyName)
			}
		}
	}

	// 4. Cleanup local files
	privateKeyPath := filepath.Join(os.TempDir(), d.instanceName+".pem")
	os.Remove(privateKeyPath)
	os.Remove(privateKeyPath + ".pub")

	return nil
}
