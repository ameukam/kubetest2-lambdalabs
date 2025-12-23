package main

import (
	"sigs.k8s.io/kubetest2/pkg/app"

	"github.com/ameukam/kubetest2-lambdalabs/deployer"
)

func main() {
	app.Main(deployer.Name, deployer.New)
}
