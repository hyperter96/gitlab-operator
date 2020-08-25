package utils

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

// CommandClient struct defines the client
// used to interact with kubernetes
type CommandClient struct {
	// client-go client
	client *kubernetes.Clientset
	// kubernetes config
	config *rest.Config
}

// CommandResult holds standard
// output and standard error
type CommandResult struct {
	// standard out
	stdOut string
	// standard error
	stdError string
}

// Output returns standard out
func (r CommandResult) Output() string {
	return r.stdOut
}

// Error returns standard error
func (r CommandResult) Error() string {
	return r.stdError
}

// NewCommandClient creates a new backup client that
// can be used to interact with kubernetes API
func NewCommandClient() (CommandClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("",
			filepath.Join(os.Getenv("HOME"), ".kube", "config"))
		if err != nil {
			return CommandClient{}, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return CommandClient{}, err
	}

	return CommandClient{
		config: config,
		client: clientset,
	}, nil
}

// TaskRunnerDeployment returns the deployment resource associated with the GitLab instance
func (c CommandClient) TaskRunnerDeployment(cr *gitlabv1beta1.GitLab) (*appsv1.Deployment, error) {
	deployment := strings.Join([]string{cr.Name, "task-runner"}, "-")
	target, err := c.client.AppsV1().Deployments(cr.Namespace).Get(context.Background(), deployment, metav1.GetOptions{})
	if err != nil {
		return &appsv1.Deployment{}, err
	}

	return target, nil
}

// DeploymentPods returns list of pods associated with a given deployment
func (c CommandClient) DeploymentPods(cr *gitlabv1beta1.GitLab, obj *appsv1.Deployment) ([]corev1.Pod, error) {

	podLabels := labels.Set(obj.Spec.Template.Labels)

	pods, err := c.client.CoreV1().Pods(cr.Namespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: podLabels.AsSelector().String()},
	)
	if err != nil {
		return []corev1.Pod{}, err
	}

	return pods.Items, nil
}

// ExecuteCommandInPod runs command in pod
// it returns the standard out and errors as string
func (c CommandClient) ExecuteCommandInPod(podName, container, namespace string, command []string) (CommandResult, error) {

	options := &corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     false,
		Stderr:    true,
		Stdout:    true,
		TTY:       false,
	}

	req := c.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").VersionedParams(options, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return CommandResult{}, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		stdOut:   stdout.String(),
		stdError: stderr.String(),
	}, nil

}

// TaskRunnerBackupPod returns single pod on which to run command
func TaskRunnerBackupPod(list *corev1.PodList) string {

	for _, pod := range list.Items {
		if pod.Status.Phase == corev1.PodRunning {
			return pod.Name
		}
	}

	return ""
}
