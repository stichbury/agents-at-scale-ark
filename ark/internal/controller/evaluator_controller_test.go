/* Copyright 2025. McKinsey & Company */

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	arkv1alpha1 "mckinsey.com/ark/api/v1alpha1"
	"mckinsey.com/ark/internal/annotations"
)

var _ = Describe("Evaluator Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		evaluator := &arkv1alpha1.Evaluator{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Evaluator")
			err := k8sClient.Get(ctx, typeNamespacedName, evaluator)
			if err != nil && errors.IsNotFound(err) {
				resource := &arkv1alpha1.Evaluator{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &arkv1alpha1.Evaluator{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Evaluator")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &EvaluatorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})

	Context("mergeEvaluationMetadata", func() {
		var reconciler *EvaluatorReconciler

		BeforeEach(func() {
			reconciler = &EvaluatorReconciler{}
		})

		It("should merge query labels and annotations into evaluation metadata", func() {
			query := &arkv1alpha1.Query{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-query",
					Generation: 5,
					Labels: map[string]string{
						"n8n_workflow_name": "test-workflow",
						"environment":       "production",
					},
					Annotations: map[string]string{
						"ark.mckinsey.com/run-id":    "run-123",
						"ark.mckinsey.com/session-id": "session-456",
					},
				},
				Status: arkv1alpha1.QueryStatus{
					Phase: "done",
				},
			}

			evaluator := &arkv1alpha1.Evaluator{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-evaluator",
				},
			}

			labels, annotationsMap := reconciler.mergeEvaluationMetadata(query, evaluator)

			// Check that query labels are copied
			Expect(labels["n8n_workflow_name"]).To(Equal("test-workflow"))
			Expect(labels["environment"]).To(Equal("production"))

			// Check that required labels are set and take precedence
			Expect(labels[annotations.Evaluator]).To(Equal("test-evaluator"))
			Expect(labels[annotations.Query]).To(Equal("test-query"))
			Expect(labels[annotations.Auto]).To(Equal("true"))

			// Check that query annotations are copied
			Expect(annotationsMap["ark.mckinsey.com/run-id"]).To(Equal("run-123"))
			Expect(annotationsMap["ark.mckinsey.com/session-id"]).To(Equal("session-456"))

			// Check that required annotations are set
			Expect(annotationsMap[annotations.QueryGeneration]).To(Equal("5"))
			Expect(annotationsMap[annotations.QueryPhase]).To(Equal("done"))
		})

		It("should override query labels and annotations with required labels and annotations when keys conflict", func() {
			query := &arkv1alpha1.Query{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-query",
					Generation: 1,
					Labels: map[string]string{
						annotations.Evaluator: "wrong-evaluator",
						annotations.Query:     "wrong-query",
						annotations.Auto:       "false",
					},
					Annotations: map[string]string{
						annotations.QueryGeneration: "999",
						annotations.QueryPhase:       "wrong-phase",
					},
				},
				Status: arkv1alpha1.QueryStatus{
					Phase: "pending",
				},
			}

			evaluator := &arkv1alpha1.Evaluator{
				ObjectMeta: metav1.ObjectMeta{
					Name: "correct-evaluator",
				},
			}

			labels, annotationsMap := reconciler.mergeEvaluationMetadata(query, evaluator)

			// Required labels should take precedence
			Expect(labels[annotations.Evaluator]).To(Equal("correct-evaluator"))
			Expect(labels[annotations.Query]).To(Equal("test-query"))
			Expect(labels[annotations.Auto]).To(Equal("true"))

			// Required annotations should take precedence (using actual query.Generation and query.Status.Phase)
			Expect(annotationsMap[annotations.QueryGeneration]).To(Equal("1"))
			Expect(annotationsMap[annotations.QueryPhase]).To(Equal("pending"))
		})
	})
})
