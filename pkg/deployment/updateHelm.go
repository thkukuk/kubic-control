package deployment

import (
	"errors"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

func checkHelmUpdate(chartName, releaseName, valuesPath, namespace, hash string) (bool, error) {
	var success bool
	var message string
	if valuesPath == "" {
		success, message = tools.ExecuteCmd("helm", "template", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf",
			"--namespace", namespace)
	} else {
		success, message = tools.ExecuteCmd("helm", "template", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf",
			"-f", valuesPath,
			"--namespace", namespace)
	}

	if success != true {
		return false, errors.New(message)
	}

	newHash, _ := tools.Sha256sum_b(message)
	if hash == newHash {
		return false, nil
	}
	return true, nil
}

func UpdateHelm(chartName, releaseName, valuesPath, namespace string) error {

	var success bool
	var message string
	if namespace == "" {
		namespace = "default"
	}
	if valuesPath == "" {
		success, message = tools.ExecuteCmd("helm", "upgrade", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf",
			"--namespace", namespace)
	} else {
		success, message = tools.ExecuteCmd("helm", "upgrade", releaseName,
			chartName, "--kubeconfig=/etc/kubernetes/admin.conf",
			"-f", valuesPath,
			"--namespace", namespace)
	}

	if success != true {
		return errors.New(message)
	}

	return setHelmConfig(chartName, releaseName, valuesPath, namespace)
}
