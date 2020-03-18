package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

const updateInterval = 10

func (r *ReconcileGitlab) updateGitlabStatus(cr *gitlabv1beta1.Gitlab) {
	labels := getLabels(cr, "gitlab")

	// If the Gitlab deployment does not exist, we assume the database is still starting up
	gitlab := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: cr.Namespace, Name: labels["app.kubernetes.io/instance"]}, gitlab)
	if err != nil {
		cr.Status.Phase = "Initializing"
		SetStatus(r.client, cr)
	}

	// Wait for database to start up
	for !isDatabaseReady(cr) {
		if cr.Status.Stage == "" {
			cr.Status.Stage = "Initializing database"
			SetStatus(r.client, cr)
		}
		time.Sleep(time.Second * updateInterval)
	}

	// Check status every ten seconds and update status
	for {
		readiness, err := getReadinessStatus(cr, labels["app.kubernetes.io/instance"])
		if err != nil {
			cr.Status.Stage = "Gitlab is starting up"
			SetStatus(r.client, cr)
		}

		if readiness.WorkhorseStatus == "ok" {
			cr.Status.Phase = "Running"
			cr.Status.Stage = ""
			cr.Status.Services.Redis = getServiceHealth(readiness.RedisStatus)
			cr.Status.Services.Database = getServiceHealth(readiness.DatabaseStatus)
			SetStatus(r.client, cr)
		}

		time.Sleep(time.Second * updateInterval)
	}

}

func getReadinessStatus(cr *gitlabv1beta1.Gitlab, service string) (*ReadinessStatus, error) {
	status := &ReadinessStatus{}

	resp, err := http.Get(fmt.Sprintf("http://%s:8005/-/readiness?all=1", service))
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Read readiness status response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return status, err
		}

		err = json.Unmarshal(body, status)
	}

	return status, err
}

// Retrieve health of a subsystem
func getServiceHealth(service []ServiceStatus) string {
	if len(service) == 1 {
		return service[0].Status
	}

	for _, item := range service {
		if item.Status != "" {
			return item.Status
		}
	}

	return "unknown"
}
