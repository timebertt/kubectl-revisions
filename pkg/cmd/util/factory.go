package util

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Factory augments the Factory interface for creating controller-runtime clients.
type Factory interface {
	cmdutil.Factory
	// Client returns a new controller-runtime client.
	Client() (client.Client, error)
}

// NewFactory creates a new factory based on the given configuration.
func NewFactory(clientGetter genericclioptions.RESTClientGetter) Factory {
	return factoryImpl{
		Factory: cmdutil.NewFactory(clientGetter),
	}
}

type factoryImpl struct {
	cmdutil.Factory
}

func (f factoryImpl) Client() (client.Client, error) {
	restConfig, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// use the factory's cached discovery rest mapper instead of controller-runtime's default rest mapper
	mapper, err := f.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	return client.New(restConfig, client.Options{Mapper: mapper})
}
