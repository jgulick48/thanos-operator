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
	"sort"

	"github.com/banzaicloud/thanos-operator/pkg/resources"
	"github.com/banzaicloud/thanos-operator/pkg/resources/rule"
	"github.com/banzaicloud/thanos-operator/pkg/resources/store"
	"github.com/banzaicloud/thanos-operator/pkg/sdk/api/v1alpha1"
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const serverCertMountPath = "/etc/tls/server"
const clientCertMountPath = "/etc/tls/client"

func New(reconciler *resources.ThanosComponentReconciler) *Receive {
	return &Receive{
		ThanosComponentReconciler: reconciler,
	}
}

type Receive struct {
	*resources.ThanosComponentReconciler
}

func (q *Receive) Reconcile() (*reconcile.Result, error) {
	if q.Thanos.Spec.Query != nil {
		err := mergo.Merge(q.Thanos.Spec.Query, v1alpha1.DefaultQuery)
		if err != nil {
			return nil, err
		}
	}
	return q.ReconcileResources(
		[]resources.Resource{
			q.deployment,
			q.service,
			q.serviceMonitor,
			q.ingressHTTP,
			q.ingressGRPC,
		})
}

func (q *Receive) getLabels() resources.Labels {
	labels := resources.Labels{
		resources.NameLabel: v1alpha1.QueryName,
	}.Merge(
		q.GetCommonLabels(),
	)
	if q.Thanos.Spec.Query != nil {
		labels.Merge(q.Thanos.Spec.Query.Labels)
	}
	return labels
}

func (q *Receive) getName(suffix ...string) string {
	name := v1alpha1.QueryName
	if len(suffix) > 0 {
		name = name + "-" + suffix[0]
	}
	return q.QualifiedName(name)
}

func (q *Receive) getSvc() string {
	return fmt.Sprintf("_grpc._tcp.%s.%s.svc.cluster.local", q.getName(), q.Thanos.Namespace)
}

func (q *Receive) getMeta(name string, params ...string) metav1.ObjectMeta {
	namespace := ""
	if len(params) > 0 {
		namespace = params[0]
	}
	meta := q.GetObjectMeta(name, namespace)
	meta.Labels = q.getLabels()
	if q.Thanos.Spec.Query != nil {
		meta.Annotations = q.Thanos.Spec.Query.Annotations
	}
	return meta
}

func (q *Receive) getStoreEndpoints() []string {
	var endpoints []string
	// Discover all query instance
	if q.Thanos.Spec.QueryDiscovery {
		for _, t := range q.ThanosList {
			if t.Spec.Query != nil {
				reconciler := resources.NewThanosComponentReconciler(t.DeepCopy(), nil, nil, nil)
				svc := (&Receive{reconciler}).getSvc()
				endpoints = append(endpoints, fmt.Sprintf("--store=dnssrvnoa+%s", svc))
			}
		}
	}
	// Discover local StoreGateway
	if q.Thanos.Spec.StoreGateway != nil {
		for _, u := range store.New(q.ThanosComponentReconciler).GetServiceURLS() {
			endpoints = append(endpoints, fmt.Sprintf("--store=dnssrvnoa+%s", u))
		}
	}
	// Discover local Rule
	if q.Thanos.Spec.Rule != nil {
		for _, u := range rule.New(q.ThanosComponentReconciler).GetServiceURLS() {
			endpoints = append(endpoints, fmt.Sprintf("--store=dnssrvnoa+%s", u))
		}
	}
	// Discover StoreEndpoint aka Sidecars
	for _, endpoint := range q.StoreEndpoints {
		if url := endpoint.GetServiceURL(); url != "" {
			endpoints = append(endpoints, fmt.Sprintf("--store=%s", url))
		}
	}
	return endpoints
}

func (q *Receive) setArgs(originArgs []string) []string {
	receive := q.Thanos.Spec.Receive.DeepCopy()
	// Get args from the tags
	args := resources.GetArgs(receive)
	if receive.RemoteWriteServerCertificate != "" {
		args = append(args, fmt.Sprintf("--remote-write.server-tls-cert=%s/%s", serverCertMountPath, "tls.crt"))
		args = append(args, fmt.Sprintf("--remote-write.server-tls-key=%s/%s", serverCertMountPath, "tls.key"))
		args = append(args, fmt.Sprintf("--remote-write.server-tls-ca=%s/%s", serverCertMountPath, "ca.crt"))
	}
	if receive.RemoteWriteClientCertificate != "" {
		args = append(args, fmt.Sprintf("--remote-write.client-tls-cert=%s/%s", clientCertMountPath, "tls.crt"))
		args = append(args, fmt.Sprintf("--remote-write.client-tls-key=%s/%s", clientCertMountPath, "tls.key"))
		args = append(args, fmt.Sprintf("--remote-write.client-tls-ca=%s/%s", clientCertMountPath, "ca.crt"))
		args = append(args, "--remote-write.client-server-name") //TODO this is dummy now
	}
	// Add discovery args
	args = append(args, q.getStoreEndpoints()...)

	// Sort generated args to prevent accidental diffs
	sort.Strings(args)
	// Concat original and computed args
	finalArgs := append(originArgs, args...)
	return finalArgs
}
