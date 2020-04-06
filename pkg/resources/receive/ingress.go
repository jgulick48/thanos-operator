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
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (q *Receive) ingressHTTP() (runtime.Object, reconciler.DesiredState, error) {
	if q.Thanos.Spec.Receive != nil &&
		q.Thanos.Spec.Receive.HTTPIngress != nil {
		receiveIngress := q.Thanos.Spec.Receive.HTTPIngress
		ingress := &v1beta1.Ingress{
			ObjectMeta: q.getMeta(q.getName("http")),
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: receiveIngress.Host,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: receiveIngress.Path,
										Backend: v1beta1.IngressBackend{
											ServiceName: q.getName(),
											ServicePort: intstr.IntOrString{
												Type:   intstr.String,
												StrVal: "http",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		if receiveIngress.Certificate != "" {
			ingress.Spec.TLS = []v1beta1.IngressTLS{
				{
					Hosts:      []string{receiveIngress.Host},
					SecretName: receiveIngress.Certificate,
				},
			}
		}
		return ingress, reconciler.StatePresent, nil
	}
	delete := &corev1.Service{
		ObjectMeta: q.getMeta(q.getName("http")),
	}
	return delete, reconciler.StateAbsent, nil
}

func (q *Receive) ingressGRPC() (runtime.Object, reconciler.DesiredState, error) {
	if q.Thanos.Spec.Receive != nil &&
		q.Thanos.Spec.Receive.GRPCIngress != nil {
		receiveIngress := q.Thanos.Spec.Receive.GRPCIngress
		ingress := &v1beta1.Ingress{
			ObjectMeta: q.getMeta(q.getName("grpc")),
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: receiveIngress.Host,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: receiveIngress.Path,
										Backend: v1beta1.IngressBackend{
											ServiceName: q.getName(),
											ServicePort: intstr.IntOrString{
												Type:   intstr.String,
												StrVal: "grpc",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		if receiveIngress.Certificate != "" {
			ingress.Spec.TLS = []v1beta1.IngressTLS{
				{
					Hosts:      []string{receiveIngress.Host},
					SecretName: receiveIngress.Certificate,
				},
			}
		}
		return ingress, reconciler.StatePresent, nil
	}
	delete := &corev1.Service{
		ObjectMeta: q.getMeta(q.getName("grpc")),
	}
	return delete, reconciler.StateAbsent, nil
}
