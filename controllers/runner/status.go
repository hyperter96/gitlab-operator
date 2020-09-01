package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// WorkerPod returns the runner pod associated with the Runner object being reconciled
func WorkerPod(cr *gitlabv1beta1.Runner, client *kubernetes.Clientset) (*corev1.Pod, error) {
	pods, err := client.CoreV1().Pods(cr.Namespace).List(context.TODO(), metav1.ListOptions{
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

// LogStream returns logs on the target pods console
func LogStream(pod *corev1.Pod, client *kubernetes.Clientset) (string, error) {
	logReader, err := client.CoreV1().RESTClient().Get().
		Resource("pods").SubResource("log").Namespace(pod.Namespace).Name(pod.Name).
		VersionedParams(&corev1.PodLogOptions{}, scheme.ParameterCodec).Stream(context.TODO())
	if err != nil {
		return "", err
	}

	logs, err := ioutil.ReadAll(logReader)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// RegistrationStatus returns status of runner registration attempt
func RegistrationStatus(log string) string {
	re := regexp.MustCompile("(?m)Registering runner([.]+) ([a-z]+)")
	registration := re.FindSubmatch([]byte(log))
	return string(registration[len(registration)-1])
}
