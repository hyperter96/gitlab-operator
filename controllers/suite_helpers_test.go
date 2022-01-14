package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

var (
	emptyValues = resource.Values{}
)

func createObject(obj client.Object, ignoreAlreadyExists bool) error {
	err := k8sClient.Create(ctx, obj)
	if errors.IsAlreadyExists(err) && ignoreAlreadyExists {
		err = nil
	}

	return err
}

func updateObject(obj client.Object, mutate func(client.Object) error) error {
	key := client.ObjectKeyFromObject(obj)
	if err := k8sClient.Get(ctx, key, obj); err != nil {
		return err
	}

	if err := mutate(obj); err != nil {
		return err
	}

	return k8sClient.Update(ctx, obj)
}

func getObject(name string, obj client.Object) error {
	lookupKey := types.NamespacedName{
		Name:      name,
		Namespace: Namespace,
	}

	return k8sClient.Get(ctx, lookupKey, obj)
}

func getObjectPromise(name string, obj client.Object) func() error {
	return func() error {
		return getObject(name, obj)
	}
}

func listObjects(query string, obj client.ObjectList) error {
	labelSelector, err := labels.Parse(query)
	if err != nil {
		return err
	}

	listOptions := &client.ListOptions{
		Namespace:     Namespace,
		LabelSelector: labelSelector,
	}

	if err := k8sClient.List(ctx, obj, listOptions); err != nil {
		return err
	}

	return nil
}

func listObjectsPromise(query string, obj client.ObjectList, expectedSize int) func() error {
	return func() error {
		if err := listObjects(query, obj); err != nil {
			return err
		}

		switch l := obj.(type) {
		case *batchv1.JobList:
			if len(l.Items) < expectedSize {
				return fmt.Errorf("Only %d Jobs found with [%s]. Expecting %d",
					len(l.Items), query, expectedSize)
			}
		}

		return nil
	}
}

func deleteObject(name string, obj client.Object) error {
	if err := getObject(name, obj); err != nil {
		if errors.IsNotFound(err) {
			err = nil
		}

		return err
	}

	return k8sClient.Delete(ctx, obj)
}

func deleteObjectPromise(name string, obj client.Object) func() error {
	return func() error {
		return deleteObject(name, obj)
	}
}

func listConfigMapsPromise(query string) func() []corev1.ConfigMap {
	return func() []corev1.ConfigMap {
		createdCfgMaps := &corev1.ConfigMapList{}
		if err := listObjects(query, createdCfgMaps); err != nil {
			return nil
		}

		return createdCfgMaps.Items
	}
}

func updateJobStatusPromise(query string, success bool) func() error {
	return func() error {
		createdJobs := &batchv1.JobList{}

		if err := listObjects(query, createdJobs); err != nil {
			return err
		}

		if len(createdJobs.Items) == 0 {
			return fmt.Errorf("Job not found [%s]", query)
		}

		for _, j := range createdJobs.Items {
			j := j
			if success {
				j.Status.Succeeded = 1
				j.Status.Failed = 0
			} else {
				j.Status.Succeeded = 0
				j.Status.Failed = 1
			}

			if err := k8sClient.Status().Update(ctx, &j); err != nil {
				return err
			}
		}

		return nil
	}
}

func appLabels(releaseName, appName string) string {
	return fmt.Sprintf("release=%s,app=%s", releaseName, appName)
}

func createGitLabResource(releaseName string, chartValues resource.Values) {
	By("Creating a new GitLab resource")
	Expect(createObject(gitlab.CreateMockGitLab(releaseName, Namespace, chartValues), true)).Should(Succeed())

	By("Checking GitLab resource is created")
	Eventually(getObjectPromise(releaseName, &gitlabv1beta1.GitLab{}),
		PollTimeout, PollInterval).Should(Succeed())
}

func updateGitLabResource(releaseName string, chartValues resource.Values) {
	By("Update the existing GitLab resource")
	Expect(
		updateObject(
			gitlab.CreateMockGitLab(releaseName, Namespace, resource.Values{}),
			func(obj client.Object) error {
				gitlab := obj.(*gitlabv1beta1.GitLab)
				gitlab.Spec.Chart.Values.Object = chartValues
				return nil
			})).Should(Succeed())
}
