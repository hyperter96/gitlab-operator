package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

func runnerPod(cr *gitlabv1beta1.Runner, client *kubernetes.Clientset) (*corev1.Pod, error) {
	pods, err := client.CoreV1().Pods(cr.Namespace).List(metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/part-of",
	})
	if err != nil {
		return nil, err
	}

	podPattern := regexp.MustCompile(fmt.Sprintf("^%s-runner-([a-z0-9]+)-[a-z0-9]+", cr.Name))
	target := corev1.Pod{}
	for _, pod := range pods.Items {
		if podPattern.MatchString(pod.Name) {
			target = pod
		}
	}

	return &target, err
}

func openLogStream(pod *corev1.Pod, client *kubernetes.Clientset) (string, error) {
	logReader, err := client.CoreV1().RESTClient().Get().
		Resource("pods").SubResource("log").Namespace(pod.Namespace).Name(pod.Name).
		VersionedParams(&corev1.PodLogOptions{}, scheme.ParameterCodec).Stream()
	if err != nil {
		return "", err
	}

	logs, err := ioutil.ReadAll(logReader)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

func runnerRegistrationStatus(log string) string {
	re := regexp.MustCompile("(?m)Registering runner([.]+) ([a-z]+)")
	registration := re.FindSubmatch([]byte(log))
	return string(registration[len(registration)-1])
}

func (r *ReconcileRunner) updateRunnerStatus(cr *gitlabv1beta1.Runner, status string) error {
	runner := &gitlabv1beta1.Runner{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, runner)
	if err != nil {
		return err
	}

	runner.Status.Phase = "Running"
	runner.Status.Registration = status

	return r.client.Status().Update(context.TODO(), runner)
}

func (r *ReconcileRunner) reconcileRunnerStatus(cr *gitlabv1beta1.Runner) error {

	client, err := gitlabutils.NewKubernetesClient()
	if err != nil {
		return err
	}

	pod, err := runnerPod(cr, client)
	if err != nil {
		return err
	}

	var log string
	if gitlabutils.IsPodRunning(pod) {
		log, err = openLogStream(pod, client)
		if err != nil {
			return err
		}
	}

	if log != "" {
		err := r.updateRunnerStatus(cr, runnerRegistrationStatus(log))
		if err != nil {
			return err
		}
	}

	return nil
}
