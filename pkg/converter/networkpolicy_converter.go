// Copyright (c) 2017 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"fmt"

	api "github.com/projectcalico/libcalico-go/lib/apis/v2"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s/conversion"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type policyConverter struct {
}

// NewPolicyConverter Constructor for policyConverter
func NewPolicyConverter() Converter {
	return &policyConverter{}
}

func (p *policyConverter) Convert(k8sObj interface{}) (interface{}, error) {
	np, ok := k8sObj.(*v1beta1.NetworkPolicy)

	if !ok {
		tombstone, ok := k8sObj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, fmt.Errorf("couldn't get object from tombstone %+v", k8sObj)
		}
		np, ok = tombstone.Obj.(*v1beta1.NetworkPolicy)
		if !ok {
			return nil, fmt.Errorf("tombstone contained object that is not a NetworkPolicy %+v", k8sObj)
		}
	}

	var c conversion.Converter
	kvp, err := c.NetworkPolicyToPolicy(np)
	if err != nil {
		return nil, err
	}
	calicoPolicy := kvp.Value.(*api.NetworkPolicy)

	// To ease upgrade path, create an allow-all Egress rule, but with Types: Ingress
	// In the case where there's an older Felix interoperating with a new kube-controllers
	// controller, Felix will respect the egress rule and ignore the types field.
	// When Felix is upgraded, it will ignore the Egress allow-all rule due to
	// Types: Ingress.
	if len(calicoPolicy.Spec.Types) == 1 && calicoPolicy.Spec.Types[0] == api.PolicyTypeIngress {
		calicoPolicy.Spec.EgressRules = []api.Rule{{Action: "allow"}}
	}
	return *calicoPolicy, err
}

// GetKey returns name of Policy as its key.  For Policies created by this controller
// and backed by NetworkPolicy objects, the name is of the format
// `knp.default.namespace.name`.
func (p *policyConverter) GetKey(obj interface{}) string {
	policy := obj.(api.NetworkPolicy)
	k, _ := cache.MetaNamespaceKeyFunc(policy)
	return k
}
