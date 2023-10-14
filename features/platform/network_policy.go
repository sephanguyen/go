package platform

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	cachedImagePrefix       string = "kind-reg.actions-runner-system.svc/"
	istioSidecarInjectLabel string = "sidecar.istio.io/inject"
	appLabel                       = "app"

	nginxDockerImage  = cachedImagePrefix + "nginx:1.25.1-alpine3.17-slim"
	alpineDockerImage = cachedImagePrefix + "alpine:3.18.2"
)

var errJobFailed = errors.New("job failed")

var networkPolicyIsEnabledInCurrentKubernetesCluster = func() func(ctx context.Context) (context.Context, error) {
	// We do some optimization here to make this check runs only once
	// for the entire suite.
	var once sync.Once
	var enabled bool
	var outerErr error
	return func(ctx context.Context) (context.Context, error) {
		s := stepStateFromContext(ctx)
		once.Do(func() {
			subctx, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()

			const calicoSystemNamespace = "calico-system"
			nsList, err := k8sClient().CoreV1().Namespaces().List(subctx, metav1.ListOptions{
				FieldSelector: "metadata.name=" + calicoSystemNamespace,
			})
			if err != nil {
				outerErr = fmt.Errorf("failed to list namespaces: %s", err)
				return
			}
			if len(nsList.Items) == 0 {
				ctxzap.Warn(ctx, "network policy is disabled, some tests might be skipped")
				return
			}

			dsList, err := k8sClient().AppsV1().DaemonSets(calicoSystemNamespace).List(subctx, metav1.ListOptions{
				FieldSelector: "metadata.name=csi-node-driver",
			})
			if err != nil {
				outerErr = fmt.Errorf("failed to list daemonsets in namespace %q: %s", calicoSystemNamespace, err)
				return
			}
			if len(dsList.Items) == 0 {
				ctxzap.Warn(ctx, "network policy is disabled, some tests might be skipped")
				return
			}

			enabled = true
			ctxzap.Info(ctx, "network policy is enabled")
		})
		s.networkPolicyEnabled = enabled
		return stepStateToContext(ctx, s), outerErr
	}
}()

func translateNamespace(org, name string) string {
	switch name {
	case "backend", "elastic", "kafka", "nats-jetstream", "unleash":
		return fmt.Sprintf("local-%s-%s", org, name)
	default:
		return name
	}
}

func podInNamespaceAccessPodInNamespace(ctx context.Context, originNamespace, canOrCannot, targetNamespace string) (context.Context, error) {
	subctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	s := stepStateFromContext(ctx)
	if !s.networkPolicyEnabled {
		return ctx, nil
	}

	// translate the namespaces since the org value
	// cannot be determined at compile-time.
	originNamespace = translateNamespace(s.org, originNamespace)
	targetNamespace = translateNamespace(s.org, targetNamespace)

	dummyName := fmt.Sprintf("nginx-%d", rand.Int()) //nolint:gosec
	cleanup, err := createNginxPodAndSvcInNamespace(subctx, dummyName, targetNamespace)
	if err != nil {
		return ctx, fmt.Errorf("failed to create pod and service %q in namespace %q: %s",
			dummyName, targetNamespace, err)
	}
	defer cleanup()

	err = pingPodFromNamespace(subctx, originNamespace, targetNamespace, dummyName, "80")
	if canOrCannot == "can" {
		if err != nil {
			return ctx, fmt.Errorf("failed to ping pod %q in namespace %q from namespace %q: %s",
				dummyName, targetNamespace, originNamespace, err)
		}
	} else {
		if err != nil {
			if !errors.Is(err, errJobFailed) && errors.Is(err, context.DeadlineExceeded) {
				ctxzap.Warn(ctx, "unexpected error from pinging",
					zap.String("origin_ns", originNamespace), zap.String("target_ns", targetNamespace), zap.Error(err))
			}
		} else {
			return ctx, fmt.Errorf("detected successful ping to pod %q in namespace %q from namespace %q; it should not be allowed",
				dummyName, targetNamespace, originNamespace)
		}
	}
	return ctx, nil
}

func createNamespaceIfNotExists(ctx context.Context, namespaceName string) error {
	nsList, err := k8sClient().CoreV1().Namespaces().List(ctx, metav1.ListOptions{FieldSelector: "metadata.name=" + namespaceName})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %s", err)
	}
	if len(nsList.Items) == 1 {
		return nil
	}
	if len(nsList.Items) > 1 {
		return fmt.Errorf("found %d namespaces matching the criteria metadata.name=%s", len(nsList.Items), namespaceName)
	}

	// namespace does not exist, we create it
	ctxzap.Debug(ctx, "creating namespace", zap.String("ns", namespaceName))
	_, err = k8sClient().CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %s", err)
	}

	return nil
}

// createNginxPodAndSvcInNamespace creates a pod and service in the target namespace.
//
// The service then can be accessed via port 80.
func createNginxPodAndSvcInNamespace(ctx context.Context, resourceName, namespaceName string) (cleanup func(), err error) {
	const portVal int32 = 80
	cleanup = func() { // clean up function
		ctxzap.Debug(ctx, "cleaning up created pod and service",
			zap.String("resourceName", resourceName))
		// use a new context just in case the parent context had been canceled
		newctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		if err := k8sClient().CoreV1().Pods(namespaceName).Delete(newctx, resourceName, metav1.DeleteOptions{}); err != nil {
			ctxzap.Warn(ctx, "error cleaning up pod",
				zap.String("pod_name", resourceName), zap.Error(err))
		}
		if err := k8sClient().CoreV1().Services(namespaceName).Delete(newctx, resourceName, metav1.DeleteOptions{}); err != nil {
			ctxzap.Warn(ctx, "error cleaning up service",
				zap.String("service_name", resourceName), zap.Error(err))
		}
	}
	defer func() { // automatically clean up created resources if this function failed unexpectedly
		if err != nil {
			cleanup()
		}
	}()

	// ensure the target namespace exists
	if err = createNamespaceIfNotExists(ctx, namespaceName); err != nil {
		err = fmt.Errorf("namespace preparation failed: %s", err)
		return
	}

	// create pod and wait for it to be ready
	if _, err = k8sClient().CoreV1().Pods(namespaceName).Create(ctx, &corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        resourceName,
			Labels:      map[string]string{appLabel: resourceName},
			Annotations: map[string]string{istioSidecarInjectLabel: "false"}, // disable istio sidecar injection
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: nginxDockerImage, Ports: []corev1.ContainerPort{{ContainerPort: portVal}}},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
			DNSPolicy:     corev1.DNSClusterFirst,
		},
	}, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("failed to create pod: %s", err)
		return
	}
	if err = waitResourceReady(ctx, namespaceName, resourceName, "pod"); err != nil {
		err = fmt.Errorf("failed to wait for pod to ready: %s", err)
		return
	}

	// create service
	if _, err = k8sClient().CoreV1().Services(namespaceName).Create(ctx, &corev1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: resourceName},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Protocol: corev1.ProtocolTCP, Port: portVal, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: portVal}},
			},
			Selector: map[string]string{appLabel: resourceName},
		},
	}, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("failed to create service: %s", err)
		return
	}
	return
}

// waitResourceReady waits until resource is ready. resourceKind can be "pod" or "job".
func waitResourceReady(ctx context.Context, namespaceName, resourceName, resourceKind string) error {
	var isReady func(ctx context.Context, namespaceName, resourceName string) (bool, error)
	switch strings.ToLower(resourceKind) {
	case "pod":
		isReady = func(ctx context.Context, namespaceName, resourceName string) (bool, error) {
			pod, err := k8sClient().CoreV1().Pods(namespaceName).Get(ctx, resourceName, metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("failed to get pod: %s", err)
			}
			for _, c := range pod.Status.Conditions {
				if c.Status == corev1.ConditionTrue {
					ctxzap.Debug(ctx, "waiting for pod readiness",
						zap.String("pod_name", resourceName), zap.String("pod_status", string(c.Type)))
					if c.Type == corev1.PodReady {
						return true, nil
					}
				}
			}
			return false, nil
		}
	case "job":
		isReady = func(ctx context.Context, namespaceName, resourceName string) (bool, error) {
			job, err := k8sClient().BatchV1().Jobs(namespaceName).Get(ctx, resourceName, metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("failed to get job: %s", err)
			}
			for _, c := range job.Status.Conditions {
				if c.Status == corev1.ConditionTrue {
					ctxzap.Debug(ctx, "waiting for job completion",
						zap.String("job_name", resourceName), zap.Reflect("job_status", job.Status))
					if c.Type == batchv1.JobComplete {
						return true, nil
					}
					if c.Type == batchv1.JobFailed {
						return false, errJobFailed
					}
				}
			}
			return false, nil
		}
	default:
		return fmt.Errorf("invalid resource kind: %q (allowed: pod, job)", resourceKind)
	}
	ticker := time.NewTicker(time.Second * 4)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ok, err := isReady(ctx, namespaceName, resourceName)
			if err != nil {
				return errors.Wrap(err, "failed to check resource readiness")
			}
			if ok {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// pingPodFromNamespace creates a k8s job in `originNamespace` namespace to ping pod `targetPod` in
// namespace `targetNamespace`.
//
// For jobs, it's not necessary to clean them up.
func pingPodFromNamespace(ctx context.Context, originNamespace, targetNamespace, targetService, targetPort string) error {
	// ensure the namespace where we create job exists
	if err := createNamespaceIfNotExists(ctx, originNamespace); err != nil {
		return fmt.Errorf("namespace preparation failed: %s", err)
	}

	jobName := fmt.Sprintf("gandalf-pinger-%d", rand.Int()) //nolint:gosec
	ctxzap.Debug(ctx, "creating job to ping",
		zap.String("job_name", jobName),
		zap.String("origin_ns", originNamespace), zap.String("target_ns", targetNamespace),
		zap.String("svc", targetService), zap.String("port", targetPort))
	var jobTTL int32 = 120
	var backoffLimit int32 = 1
	j := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{Kind: "Job", APIVersion: "batch/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: originNamespace,
		},
		Spec: batchv1.JobSpec{
			Selector: &metav1.LabelSelector{},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{istioSidecarInjectLabel: "false"}, // disable istio sidecar injection
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "gandalf-pinger",
						Image: alpineDockerImage,
						Command: []string{
							"/bin/sh", "-c",
							fmt.Sprintf("set -eu\nwget -qO- --timeout 10 http://%s.%s.svc:%s/", targetService, targetNamespace, targetPort),
						},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &jobTTL,
		},
	}
	_, err := k8sClient().BatchV1().Jobs(originNamespace).Create(ctx, j, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %s", err)
	}
	if err := waitResourceReady(ctx, originNamespace, jobName, "job"); err != nil {
		return errors.Wrap(err, "failed to wait for job completion")
	}
	return nil
}
