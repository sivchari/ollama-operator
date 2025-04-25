package envtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	ollamav1alpha1 "github.com/sivchari/ollama-operator/api/v1alpha1"
)

func init() {
	logOptions := logs.NewOptions()
	logOptions.Verbosity = logsv1.VerbosityLevel(2)
	if err := logsv1.ValidateAndApply(logOptions, nil); err != nil {
		klog.ErrorS(err, "Unable to validate and apply log options")
		os.Exit(1)
	}
	logger := klog.Background()
	ctrl.SetLogger(logger)
	klog.SetOutput(ginkgo.GinkgoWriter)
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))
	utilruntime.Must(ollamav1alpha1.AddToScheme(scheme.Scheme))
}

type Input struct {
	M                *testing.M
	SetupReconcilers func(ctx context.Context, mgr ctrl.Manager)
	SetupEnv         func(e *Environment)
}

func Run(ctx context.Context, input Input) int {
	env := newEnvironment()
	ctx, cancel := context.WithCancel(ctx)
	env.cancelFunc = cancel

	if input.SetupEnv != nil {
		input.SetupEnv(env)
	}

	if input.SetupReconcilers != nil {
		input.SetupReconcilers(ctx, env.manager)
	}

	env.start(ctx)

	defer func() {
		if err := env.stop(); err != nil {
			klog.Fatalf("failed to stop envtest: %v", err)
		}
	}()

	klog.V(1).Info("starting envtest")

	return input.M.Run()
}

type Environment struct {
	client.Client
	manager    ctrl.Manager
	env        *envtest.Environment
	cancelFunc context.CancelFunc
}

func newEnvironment() *Environment {
	env := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	if _, err := env.Start(); err != nil {
		klog.Fatalf("failed to start envtest: %v", err)
	}

	mgr, err := ctrl.NewManager(env.Config, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		klog.Fatalf("unable to start manager: %v", err)
	}

	return &Environment{
		Client:  mgr.GetClient(),
		manager: mgr,
		env:     env,
	}
}

func (e *Environment) start(ctx context.Context) {
	go func() {
		if err := e.manager.Start(ctx); err != nil {
			klog.Fatalf("failed to start manager: %v", err)
		}
	}()
	<-e.manager.Elected()
}

func (e *Environment) stop() error {
	klog.V(1).Info("stopping envtest")
	e.cancelFunc()
	return e.env.Stop()
}

func (e *Environment) CreateNamespace(ctx context.Context, ns string) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", ns),
		},
	}
	if err := e.Create(ctx, namespace); err != nil {
		return nil, err
	}
	return namespace, nil
}
