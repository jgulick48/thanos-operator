// Copyright 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package receive

import (
	"fmt"

	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/thanos-operator/pkg/resources"
)

func (q *Receive) deployment() (runtime.Object, reconciler.DesiredState, error) {
	if q.Thanos.Spec.Receive != nil {
		receive := q.Thanos.Spec.Receive.DeepCopy()
		var deployment = &appsv1.Deployment{
			ObjectMeta: q.getMeta(q.getName()),
			Spec: appsv1.DeploymentSpec{
				Replicas: utils.IntPointer(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: q.getLabels(),
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: q.getMeta(q.getName()),
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "receive",
								Image: fmt.Sprintf("%s:%s", receive.Image.Repository, receive.Image.Tag),
								Args: []string{
									"receive",
								},
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										ContainerPort: resources.GetPort(receive.HttpAddress),
										Protocol:      corev1.ProtocolTCP,
									},
									{
										Name:          "grpc",
										ContainerPort: resources.GetPort(receive.GRPCAddress),
										Protocol:      corev1.ProtocolTCP,
									},
								},
								Resources:       receive.Resources,
								ImagePullPolicy: receive.Image.PullPolicy,
								LivenessProbe:   q.GetCheck(resources.GetPort(receive.HttpAddress), resources.HealthCheckPath),
								ReadinessProbe:  q.GetCheck(resources.GetPort(receive.HttpAddress), resources.ReadyCheckPath),
								VolumeMounts:    q.getVolumeMounts(),
							},
						},
						Volumes: q.getVolumes(),
					},
				},
			},
		}
		// Set up args
		deployment.Spec.Template.Spec.Containers[0].Args = q.setArgs(deployment.Spec.Template.Spec.Containers[0].Args)
		return deployment, reconciler.StatePresent, nil
	}
	delete := &appsv1.Deployment{
		ObjectMeta: q.getMeta(q.getName()),
	}
	return delete, reconciler.StateAbsent, nil
}

func (q *Receive) getVolumeMounts() []corev1.VolumeMount {
	volumeMounts := make([]corev1.VolumeMount, 0)
	if q.Thanos.Spec.Receive.RemoteWriteClientCertificate != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "client-certificate",
			ReadOnly:  true,
			MountPath: clientCertMountPath,
		})
	}
	if q.Thanos.Spec.Receive.RemoteWriteServerCertificate != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "server-certificate",
			ReadOnly:  true,
			MountPath: serverCertMountPath,
		})
	}
	return volumeMounts
}

func (q *Receive) getVolumes() []corev1.Volume {
	volumes := make([]corev1.Volume, 0)
	if q.Thanos.Spec.Receive.RemoteWriteClientCertificate != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "client-certificate",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: q.Thanos.Spec.Query.GRPCClientCertificate,
				},
			},
		})
	}
	if q.Thanos.Spec.Receive.RemoteWriteServerCertificate != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "server-certificate",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: q.Thanos.Spec.Query.GRPCServerCertificate,
				},
			},
		})
	}
	return volumes
}
