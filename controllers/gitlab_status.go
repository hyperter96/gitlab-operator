package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/prometheus/common/log"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// EndpointMembers returns a list of members
var EndpointMembers []string

func (r *GitLabReconciler) reconcileGitlabStatus(cr *gitlabv1beta1.GitLab) error {
	// get current Gitlab resource
	gitlab, err := r.retrieveUpdatedGitlabResource(cr)
	if err != nil {
		return err
	}

	// Check if the postgres statefulset exists
	if r.isPostgresDeployed(cr) && !r.isWebserviceDeployed(cr) {
		gitlab.Status.Phase = "Initializing"
		gitlab.Status.Stage = "Waiting for database"
	}

	// Check if the webservice deployment exists
	if r.isWebserviceDeployed(cr) {
		// Find webservice pod(s)
		pods := r.getEndpointMembers(cr, cr.Name+"-webservice")
		if len(pods) == 0 {
			gitlab.Status.Phase = "Initializing"
			gitlab.Status.Stage = "Gitlab is initializing"
		}

		if len(pods) > 0 {
			gitlab.Status.Phase = "Running"
			gitlab.Status.Stage = ""

			gitlab.Status.HealthCheck = getReadinessStatus(cr)
		}
	}

	// Check if the status of the gitlab resource has changed
	if !reflect.DeepEqual(cr.Status, gitlab.Status) {
		// Update status if the status has changed
		if err := r.setGitlabStatus(gitlab); err != nil {
			return err
		}
	}

	return nil
}

func getReadinessStatus(cr *gitlabv1beta1.GitLab) *gitlabv1beta1.HealthCheck {
	var err error
	status := &ReadinessStatus{}

	resp, err := http.Get(fmt.Sprintf("http://%s:8181/-/readiness?all=1", cr.Name+"-webservice"))
	if err != nil {
		// log.Error(err, "Unable to retrieve status")
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Read readiness status response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// log.Error(err, "Unable to read status")
			return nil
		}

		if err = json.Unmarshal(body, status); err != nil {
			// log.Error(err, "Unable to convert response to struct")
			return nil
		}
	}

	return parseStatus(status)
}

// ReadinessStatus shows status of Gitlab services
type ReadinessStatus struct {
	// Returns status of Gitlab rails app
	WorkhorseStatus string `json:"status,omitempty"`
	// RedisStatus reports status of redis
	RedisStatus []ServiceStatus `json:"redis_check,omitempty"`
	// DatabaseStatus reports status of postgres
	DatabaseStatus []ServiceStatus `json:"db_check,omitempty"`
}

// ServiceStatus shows status of a Gitlab
// dependent service .e.g. Postgres, Redis, Gitaly
type ServiceStatus struct {
	Status string `json:"status,omitempty"`
}

// Retrieve health of a subsystem
func parseStatus(status *ReadinessStatus) *gitlabv1beta1.HealthCheck {
	var result gitlabv1beta1.HealthCheck

	if status.WorkhorseStatus != "" {
		result.Workhorse = status.WorkhorseStatus
	}
	// Get redis status
	if len(status.RedisStatus) == 1 {
		result.Redis = status.RedisStatus[0].Status
	}

	// Get postgresql status
	if len(status.DatabaseStatus) == 1 {
		result.Postgres = status.DatabaseStatus[0].Status
	}

	return &result
}

func (r *GitLabReconciler) isPostgresDeployed(cr *gitlabv1beta1.GitLab) bool {
	labels := gitlabutils.Label(cr.Name, "postgresql", gitlabutils.GitlabType)

	postgres := &appsv1.StatefulSet{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: cr.Namespace, Name: labels["app.kubernetes.io/instance"]}, postgres)
	return !reflect.DeepEqual(*postgres, appsv1.Deployment{}) || !errors.IsNotFound(err)
}

func (r *GitLabReconciler) isWebserviceDeployed(cr *gitlabv1beta1.GitLab) bool {
	webservice := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-webservice", Namespace: cr.Namespace}, webservice)
	return !reflect.DeepEqual(*webservice, appsv1.Deployment{}) || !errors.IsNotFound(err)
}

func (r *GitLabReconciler) retrieveUpdatedGitlabResource(cr *gitlabv1beta1.GitLab) (*gitlabv1beta1.GitLab, error) {
	gitlab := &gitlabv1beta1.GitLab{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: cr.Namespace, Name: cr.Name}, gitlab)
	return gitlab, err
}

// setGitlabStatus sets status of custom resource
func (r *GitLabReconciler) setGitlabStatus(object runtime.Object) error {
	return r.Status().Update(context.TODO(), object)
}

func (r *GitLabReconciler) getEndpointMembers(cr *gitlabv1beta1.GitLab, endpoint string) []string {
	members := []string{}

	ep := &corev1.Endpoints{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: endpoint, Namespace: cr.Namespace}, ep)
	if err != nil {
		log.Error(err, "Error getting endpoints")
	}

	for _, subset := range ep.Subsets {
		// Get members that are ready
		for _, addr := range subset.Addresses {
			members = append(members, addr.TargetRef.Name)
		}
	}

	return members
}
