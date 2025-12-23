package deployer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"k8s.io/klog/v2"
)

func (d *deployer) Up() error {
	if d.apiKey == "" {
		return fmt.Errorf("Lambda AI API key must be provided via --api-key flag or LAMBDA_AI_API_KEY env var")
	}

	// 0. Make instance name unique
	d.instanceName = fmt.Sprintf("%s-%s", d.instanceName, time.Now().Format(time.RFC3339))

	// 1. Setup SSH Key
	privateKeyPath := filepath.Join(os.TempDir(), d.instanceName+".pem")
	if d.sshKeyName == "" {
		klog.Info("No SSH key provided, generating one...")
		d.sshKeyName = d.instanceName

		// Generate key pair
		pubKeyBytes, err := generateSSHKey(privateKeyPath)
		if err != nil {
			return err
		}

		// Add to Lambda
		// Check if it exists first?
		keys, err := d.listSSHKeys()
		if err != nil {
			return fmt.Errorf("failed to list ssh keys: %v", err)
		}

		exists := false
		for _, k := range keys {
			if k.Name == d.sshKeyName {
				exists = true
				break
			}
		}

		if !exists {
			if err := d.addSSHKey(d.sshKeyName, string(pubKeyBytes)); err != nil {
				return fmt.Errorf("failed to add ssh key to lambda: %v", err)
			}
		} else {
			klog.Infof("SSH key %s already exists, assuming matching private key at %s", d.sshKeyName, privateKeyPath)
		}
	} else {
		// User provided key name. We assume they have the private key or agent configured?
		// But for automation, we usually need the private key file.
		// If user passes --ssh-key=mykey, we don't know where the private key is.
		// For now, fail if private key not found?
		// Or assume it's in ~/.ssh/id_rsa?
		// Let's warn.
		klog.Warningf("Using existing SSH key %s. Make sure you have the private key available.", d.sshKeyName)
		// For the purpose of this implementation, we might fail later if we can't SSH.
	}

	// 2. Launch Instance
	klog.Infof("Launching instance %s (type: %s, region: %s)...", d.instanceName, d.instanceType, d.region)
	instanceID, err := d.launchInstance(d.instanceName)
	if err != nil {
		return fmt.Errorf("failed to launch instance: %v", err)
	}
	klog.Infof("Instance launched with ID: %s", instanceID)

	// 3. Wait for IP
	klog.Info("Waiting for instance to be ready and have an IP...")
	var instanceIP string
	for i := 0; i < 60; i++ { // Wait up to 10 minutes (60 * 10s)
		instances, err := d.listInstances()
		if err != nil {
			klog.Warningf("Failed to list instances: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		found := false
		for _, inst := range instances {
			if inst.ID == instanceID {
				found = true
				if inst.IP != "" {
					instanceIP = inst.IP
				}
				break
			}
		}

		if instanceIP != "" {
			break
		}
		if !found {
			klog.Warningf("Instance %s not found in list yet", instanceID)
		}
		time.Sleep(10 * time.Second)
	}

	if instanceIP == "" {
		return fmt.Errorf("timed out waiting for instance IP")
	}
	klog.Infof("Instance is ready at IP: %s", instanceIP)

	// 4. Wait for SSH
	klog.Info("Waiting for SSH to be available...")
	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=10",
		"-i", privateKeyPath,
		fmt.Sprintf("%s@%s", d.sshUser, instanceIP),
		"echo", "ssh-ready",
	}

	sshReady := false
	for i := 0; i < 60; i++ { // Wait 10 minutes
		cmd := exec.Command("ssh", sshArgs...)
		if output, err := cmd.CombinedOutput(); err == nil {
			sshReady = true
			break
		} else {
			klog.V(2).Infof("SSH not ready yet: %v, output: %s", err, string(output))
		}
		time.Sleep(10 * time.Second)
	}

	if !sshReady {
		return fmt.Errorf("timed out waiting for SSH")
	}

	klog.Infof("Instance %s is UP and SSH is available.", d.instanceName)
	return nil
}

func (d *deployer) IsUp() (up bool, err error) {
	// Simple check: if instance exists and is running?
	// Or check if kubeconfig works?
	// For now, check if instance is active.
	instances, err := d.listInstances()
	if err != nil {
		return false, err
	}
	for _, inst := range instances {
		if inst.Name == d.instanceName && inst.Status == "active" {
			return true, nil
		}
	}
	return false, nil
}

