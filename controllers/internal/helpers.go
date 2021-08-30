package internal

import (
	"crypto/sha256"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
)

// PopulateAttachedSecrets populates all the Secrets that are attached to a ReplicaSet.
// nolint:nestif,gocognit // This function is a bit complicated, but breaking it up may not increase legibility.
func PopulateAttachedSecrets(template v1.PodTemplateSpec) map[string]map[string]struct{} {
	result := map[string]map[string]struct{}{}

	// Populate volumes
	for _, v := range template.Spec.Volumes {
		if v.Secret != nil {
			bucket := result[v.Secret.SecretName]
			if bucket == nil {
				bucket = map[string]struct{}{}
				result[v.Secret.SecretName] = bucket
			}

			if len(v.Secret.Items) == 0 {
				bucket["*"] = struct{}{}
			} else {
				for _, k := range v.Secret.Items {
					if _, ok := bucket[k.Key]; ok {
						continue
					}
					bucket[k.Key] = struct{}{}
				}
			}
		} else if v.Projected != nil {
			for _, s := range v.Projected.Sources {
				if s.Secret != nil {
					bucket := result[s.Secret.Name]
					if bucket == nil {
						bucket = map[string]struct{}{}
						result[s.Secret.Name] = bucket
					}
					if len(s.Secret.Items) == 0 {
						bucket["*"] = struct{}{}
					} else {
						for _, k := range s.Secret.Items {
							if _, ok := bucket[k.Key]; ok {
								continue
							}
							bucket[k.Key] = struct{}{}
						}
					}
				}
			}
		}
	}

	// Populate environment variables
	allContainers := make([]v1.Container, len(template.Spec.InitContainers)+len(template.Spec.Containers))
	allContainers = append(allContainers, template.Spec.InitContainers...)
	allContainers = append(allContainers, template.Spec.Containers...)

	for _, c := range allContainers {
		for _, e := range c.Env {
			if e.ValueFrom == nil || e.ValueFrom.SecretKeyRef == nil {
				continue
			}

			bucket := result[e.ValueFrom.SecretKeyRef.Name]
			if bucket == nil {
				bucket = map[string]struct{}{}
				result[e.ValueFrom.SecretKeyRef.Name] = bucket
			}

			if _, ok := bucket[e.ValueFrom.SecretKeyRef.Key]; ok {
				continue
			}

			bucket[e.ValueFrom.SecretKeyRef.Key] = struct{}{}
		}

		for _, e := range c.EnvFrom {
			if e.SecretRef == nil {
				continue
			}

			bucket := result[e.SecretRef.Name]
			if bucket == nil {
				bucket = map[string]struct{}{}
				result[e.SecretRef.Name] = bucket
			}

			if _, ok := bucket["*"]; ok {
				continue
			}

			bucket["*"] = struct{}{}
		}
	}

	return result
}

// SecretChecksum returns a checksum for a given Secret.
func SecretChecksum(secret v1.Secret, keys map[string]struct{}) string {
	ks := make([]string, len(secret.Data))
	for k := range secret.Data {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	hash := sha256.New()
	_, useAny := keys["*"]

	for i := range ks {
		_, useItem := keys[ks[i]]
		if useAny || useItem {
			if _, err := hash.Write(secret.Data[ks[i]]); err != nil {
				return ""
			}
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}
