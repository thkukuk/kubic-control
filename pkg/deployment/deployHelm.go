package deployment

import (
	"errors"
	"github.com/thkukuk/kubic-control/pkg/tools"
	"gopkg.in/ini.v1"
)

const (
	helmConfig      = "/var/lib/kubic-control/k8s-helm.conf"
	adminKubeconfig = "/etc/kubernetes/admin.conf"
)

func setHelmConfig(chartName, releaseName, valuesPath, namespace string) error {

	var success bool
	var message string
	if valuesPath == "" {
		success, message = tools.ExecuteCmd("helm", "template", releaseName,
			chartName, "--kubeconfig="+adminKubeconfig)
	} else {
		success, message = tools.ExecuteCmd("helm", "template", releaseName,
			chartName, "--kubeconfig="+adminKubeconfig,
			"-f", valuesPath)
	}

	if success != true {
		return errors.New(message)
	}

	result, err := tools.Sha256sum_b(message)

	cfg, err := ini.LooseLoad(helmConfig)
	if err != nil {
		return err
	}

	cfg.Section("").Key(chartName).SetValue(result)
	cfg.Section("").Key(chartName + ".releaseName").SetValue(releaseName)
	cfg.Section("").Key(chartName + ".valuesPath").SetValue(valuesPath)
	cfg.Section("").Key(chartName + ".namespace").SetValue(namespace)

	err = cfg.SaveTo(helmConfig)
	if err != nil {
		return err
	}
	return nil
}

func DeployHelm(chartName, releaseName, valuesPath, namespace string) error {

	var success bool
	var message string
	if namespace == "" {
		namespace = "default"
	}
	if valuesPath == "" {
		success, message = tools.ExecuteCmd("helm", "install", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf",
			"--namespace", namespace)
	} else {
		success, message = tools.ExecuteCmd("helm", "install", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf",
			"-f", valuesPath,
			"--namespace", namespace)
	}

	if success != true {
		return errors.New(message)
	}

	return setHelmConfig(chartName, releaseName, valuesPath, namespace)
}
