package deployment

import (
	"errors"
        "gopkg.in/ini.v1"
        "github.com/thkukuk/kubic-control/pkg/tools"
)

const(
	helmConfig = "/var/lib/kubic-control/k8s-helm.conf"
	adminKubeconfig = "/etc/kubernetes/admin.conf"
)

func setHelmConfig(chartName, releaseName, valuesPath string) error{

	var success bool
	var message string
	if valuesPath == ""{
		success, message = tools.ExecuteCmd("helm", "template", releaseName,
			chartName, "--kubeconfig="+adminKubeconfig)
	}else{
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
	cfg.Section("").Key(chartName+".releaseName").SetValue(releaseName)
	cfg.Section("").Key(chartName+".valuesPath").SetValue(valuesPath)

	
	err = cfg.SaveTo(helmConfig)
        if err != nil {
		return err
        }
	return nil
}

func DeployHelm(chartName, releaseName, valuesPath string) error {

	var success bool
	var message string
	if valuesPath == ""{
		success, message = tools.ExecuteCmd("helm", "install", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf")
	}else{
		success, message = tools.ExecuteCmd("helm", "install", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf",
			"-f", valuesPath)
	}
	
	if success != true {
		return errors.New(message)
	}

	return setHelmConfig(chartName, releaseName, valuesPath)
}
