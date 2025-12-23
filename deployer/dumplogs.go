package deployer

import "errors"

func (d *deployer) DumpClusterLogs() error {
	return errors.New("this deployer does not support dumping logs")
}
