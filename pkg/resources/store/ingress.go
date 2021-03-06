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

package store

import (
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (e *storeInstance) ingressGRPC() (runtime.Object, reconciler.DesiredState, error) {
	if e.Spec.Ingress != nil {
		endpointIngress := e.Spec.Ingress
		ingress := &v1beta1.Ingress{
			ObjectMeta: e.getMeta(),
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: endpointIngress.Host,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: endpointIngress.Path,
										Backend: v1beta1.IngressBackend{
											ServiceName: e.GetName(),
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
		if endpointIngress.Certificate != "" {
			ingress.Spec.TLS = []v1beta1.IngressTLS{
				{
					Hosts:      []string{endpointIngress.Host},
					SecretName: endpointIngress.Certificate,
				},
			}
		}
		return ingress, reconciler.StatePresent, nil
	}
	delete := &v1beta1.Ingress{
		ObjectMeta: e.getMeta(),
	}
	return delete, reconciler.StateAbsent, nil
}
