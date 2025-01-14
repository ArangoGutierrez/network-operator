/*
2023 NVIDIA CORPORATION & AFFILIATES

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers //nolint:dupl

import (
	goctx "context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mellanoxv1alpha1 "github.com/Mellanox/network-operator/api/v1alpha1"
	"github.com/Mellanox/network-operator/pkg/consts"
	"github.com/Mellanox/network-operator/pkg/nodeinfo"
)

//nolint:dupl
var _ = Describe("NicClusterPolicyReconciler Controller", func() {

	Context("When NicClusterPolicy CR is created", func() {
		It("should create whereabouts and delete it after un-setting CR value", func() {
			By("Check NicClusterPolicy with whereabouts")
			cr := mellanoxv1alpha1.NicClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nic-cluster-policy",
					Namespace: "",
				},
				Spec: mellanoxv1alpha1.NicClusterPolicySpec{
					SecondaryNetwork: &mellanoxv1alpha1.SecondaryNetworkSpec{
						IpamPlugin: &mellanoxv1alpha1.ImageSpec{
							Image:            "whereabouts",
							Repository:       "ghcr.io/k8snetworkplumbingwg",
							Version:          "v0.5.4-amd64",
							ImagePullSecrets: []string{},
						},
					},
				},
			}

			err := k8sClient.Create(goctx.TODO(), &cr)
			Expect(err).NotTo(HaveOccurred())

			ncp := &mellanoxv1alpha1.NicClusterPolicy{}
			err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: cr.GetNamespace(), Name: cr.GetName()}, ncp)
			Expect(err).NotTo(HaveOccurred())

			By("Check DS created with state label")
			Eventually(func() bool {
				ds := &appsv1.DaemonSet{}
				err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: namespaceName, Name: "whereabouts"}, ds)
				if err != nil {
					return false
				}
				l, ok := ds.Labels[consts.StateLabel]
				if !ok {
					return false
				}
				return l == "state-whereabouts-cni"
			}, timeout*3, interval).Should(BeTrue())

			By("Check SA created with state label")
			Eventually(func() bool {
				ds := &corev1.ServiceAccount{}
				err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: namespaceName, Name: "whereabouts"}, ds)
				if err != nil {
					return false
				}
				l, ok := ds.Labels[consts.StateLabel]
				if !ok {
					return false
				}
				return l == "state-whereabouts-cni"
			}, timeout*3, interval).Should(BeTrue())

			By("Update CR to remove whereabout")
			ncp = &mellanoxv1alpha1.NicClusterPolicy{}
			err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: cr.GetNamespace(), Name: cr.GetName()}, ncp)
			Expect(err).NotTo(HaveOccurred())

			ncp.Spec.SecondaryNetwork = nil
			err = k8sClient.Update(goctx.TODO(), ncp)
			Expect(err).NotTo(HaveOccurred())

			By("Check DS is deleted")
			Eventually(func() bool {
				ds := &appsv1.DaemonSet{}
				err := k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: namespaceName, Name: "whereabouts"}, ds)
				return errors.IsNotFound(err)
			}, timeout*3, interval).Should(BeTrue())

			By("Check SA is deleted")
			Eventually(func() bool {
				sa := &corev1.ServiceAccount{}
				err := k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: namespaceName, Name: "whereabouts"}, sa)
				return errors.IsNotFound(err)
			}, timeout*3, interval).Should(BeTrue())

			By("Delete NicClusterPolicy")
			err = k8sClient.Delete(goctx.TODO(), &cr)
			Expect(err).NotTo(HaveOccurred())
		})
		It("Unsupported name", func() {
			cr := mellanoxv1alpha1.NicClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "",
				},
			}
			err := k8sClient.Create(goctx.TODO(), &cr)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() string {
				found := &mellanoxv1alpha1.NicClusterPolicy{}
				err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: cr.GetNamespace(), Name: cr.GetName()}, found)
				Expect(err).NotTo(HaveOccurred())
				return string(found.Status.State)
			}, timeout*3, interval).Should(BeEquivalentTo(mellanoxv1alpha1.StateIgnore))

			err = k8sClient.Delete(goctx.TODO(), &cr)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("When NicClusterPolicy CR is deleted", func() {
		It("should set mofed.wait to false", func() {
			By("Create Node")
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-node",
					Labels:      make(map[string]string),
					Annotations: make(map[string]string),
				},
			}
			err := k8sClient.Create(goctx.TODO(), node)
			Expect(err).NotTo(HaveOccurred())
			By("Create NicClusterPolicy")
			cr := mellanoxv1alpha1.NicClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nic-cluster-policy",
					Namespace: "",
				},
				Spec: mellanoxv1alpha1.NicClusterPolicySpec{
					OFEDDriver: &mellanoxv1alpha1.OFEDDriverSpec{
						ImageSpec: mellanoxv1alpha1.ImageSpec{
							Image:            "mofed",
							Repository:       "nvcr.io/nvidia/mellanox",
							Version:          "5.9-0.5.6.0",
							ImagePullSecrets: []string{},
						},
					},
				},
			}
			err = k8sClient.Create(goctx.TODO(), &cr)
			Expect(err).NotTo(HaveOccurred())

			ncp := &mellanoxv1alpha1.NicClusterPolicy{}
			err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: cr.GetNamespace(), Name: cr.GetName()}, ncp)
			Expect(err).NotTo(HaveOccurred())

			By("Wait for NicClusterPolicy state to be populated")
			Eventually(func() string {
				found := &mellanoxv1alpha1.NicClusterPolicy{}
				err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: cr.GetNamespace(), Name: cr.GetName()}, found)
				Expect(err).NotTo(HaveOccurred())
				return string(found.Status.State)
			}, timeout*3, interval).Should(BeEquivalentTo(mellanoxv1alpha1.StateNotReady))

			By("Update Node labels")
			n := &corev1.Node{}
			err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: node.GetNamespace(), Name: node.GetName()}, n)
			Expect(err).NotTo(HaveOccurred())

			patch := []byte(fmt.Sprintf(`{"metadata":{"labels":{%q:"true", %q:"true"}}}`,
				nodeinfo.NodeLabelWaitOFED, nodeinfo.NodeLabelMlnxNIC))
			err = k8sClient.Patch(goctx.TODO(), n, client.RawPatch(types.StrategicMergePatchType, patch))
			Expect(err).NotTo(HaveOccurred())

			Consistently(func() bool {
				n := &corev1.Node{}
				err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: node.GetNamespace(), Name: node.GetName()}, n)
				if err != nil {
					return false
				}
				return n.ObjectMeta.Labels[nodeinfo.NodeLabelWaitOFED] == "true"
			}, timeout, interval).Should(BeTrue())

			By("Delete NicClusterPolicy")
			err = k8sClient.Delete(goctx.TODO(), &cr)
			Expect(err).NotTo(HaveOccurred())

			By("Verify Mofed Label is false")
			Eventually(func() bool {
				n := &corev1.Node{}
				err = k8sClient.Get(goctx.TODO(), types.NamespacedName{Namespace: node.GetNamespace(), Name: node.GetName()}, n)
				if err != nil {
					return false
				}
				return n.ObjectMeta.Labels[nodeinfo.NodeLabelWaitOFED] == "false"
			}, timeout*3, interval).Should(BeTrue())

			By("Delete Node")
			err = k8sClient.Delete(goctx.TODO(), node)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
