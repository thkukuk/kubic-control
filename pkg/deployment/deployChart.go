package deployment

import (
        "gopkg.in/ini.v1"
        "github.com/thkukuk/kubic-control/pkg/tools"
)

func DeployChart(chartName, releaseName, valuesPath string) (bool, string) {

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
		return success, message
	}

	result, err := tools.Sha256sum_f(chartName)

	cfg, err := ini.LooseLoad("/var/lib/kubic-control/k8s-yaml.conf")
	if err != nil {
		return false, "Cannot load k8s-yaml.conf: " + err.Error()
        }

	cfg.Section("").Key(chartName).SetValue(result)
	err = cfg.SaveTo("/var/lib/kubic-control/k8s-yaml.conf")
        if err != nil {
		return false, "Cannot write k8s-yaml.conf: " + err.Error()
        }

	return true, ""
}
