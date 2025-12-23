package deployer

import "k8s.io/klog/v2"

func (d *deployer) Build() error {
	klog.Info("Build is not supported/needed for this deployer, skipping.")
	return nil
}
