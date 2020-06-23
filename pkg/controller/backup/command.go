package backup

import (
	"bytes"
	"fmt"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecRemoteCommand runs command in pod
func (r *ReconcileBackup) ExecRemoteCommand(cr *gitlabv1beta1.Backup, command []string) (string, error) {
	labels := gitlabutils.Label(cr.Spec.Instance, "task-runner", gitlabutils.GitlabType)
	taskRunnerPod := r.getTaskRunnerPodName(labels["app.kubernetes.io/instance"], cr)

	config := gitlabutils.KubernetesConfig()
	client, err := config.NewKubernetesClient()
	if err != nil {
		return "", err
	}

	options := &corev1.PodExecOptions{
		Container: "task-runner",
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}

	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(taskRunnerPod).
		Namespace(cr.Namespace).
		SubResource("exec").
		VersionedParams(options, scheme.ParameterCodec)

	clientConfig := config.Config
	if err := config.Error; err != nil {
		return "", err
	}

	fmt.Println(req.URL())
	exec, err := remotecommand.NewSPDYExecutor(clientConfig, "POST", req.URL())
	if err != nil {
		return "", err
	}

	// var input io.Reader
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return stdout.String(), fmt.Errorf("Backup error: %s", stderr.String())
}

func (r *ReconcileBackup) getTaskRunnerPodName(name string, cr *gitlabv1beta1.Backup) string {

	pods, err := gitlabutils.GetDeploymentPods(r.client, name, cr.Namespace)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Deployment not found")
	}

	if len(pods) > 0 {
		return pods[0].GetName()
	}

	return ""
}
