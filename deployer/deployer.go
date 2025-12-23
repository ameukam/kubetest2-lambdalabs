package deployer

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubetest2/pkg/types"
)

const (
	Name = "Lambda AI"
)

type deployer struct {
	commonOptions types.Options

	instanceName string
	region       string
	instanceType string
	image        string
	sshKeyName   string
	sshUser      string
	apiKey       string
}

func New(opts types.Options) (types.Deployer, *pflag.FlagSet) {
	d := &deployer{
		commonOptions: opts,
		apiKey:        os.Getenv("LAMBDA_AI_API_KEY"),
	}
	return d, bindFlags(d)
}

func bindFlags(d *deployer) *pflag.FlagSet {
	flags := pflag.NewFlagSet(Name, pflag.ContinueOnError)
	flags.StringVar(&d.instanceName, "instance-name", "kubetest2-lambda", "Name of the instance to create")
	flags.StringVar(&d.region, "region", "us-east-1", "Region to deploy to")
	flags.StringVar(&d.instanceType, "instance-type", "gpu_1x_a100_sxm4", "Instance type to use")
	flags.StringVar(&d.image, "image", "Lambda Stack 24.04", "Instance image to use")
	flags.StringVar(&d.sshKeyName, "ssh-key", "", "Name of the SSH key to use")
	flags.StringVar(&d.sshUser, "ssh-user", "ubuntu", "SSH user to use")
	flags.StringVar(&d.apiKey, "api-key", d.apiKey, "Lambda AI API key (can also be set via LAMBDA_AI_API_KEY env var)")
	return flags
}

func (d *deployer) Provider() string {
	return Name
}

func (d *deployer) Kubeconfig() (string, error) {
	return "", fmt.Errorf("kubeconfig is not supported by this deployer")
}
