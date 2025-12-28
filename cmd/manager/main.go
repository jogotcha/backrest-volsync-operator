package main

import (
	"flag"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/johan/backrest-volsync-operator/api/v1alpha1"
	"github.com/johan/backrest-volsync-operator/controllers"
)

func main() {
	var metricsAddr string
	var probeAddr string
	var leaderElect bool
	var logLevel string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&leaderElect, "leader-elect", true, "Enable leader election for controller manager.")
	flag.StringVar(&logLevel, "log-level", "info", "Log level: debug|info")
	flag.Parse()

	logger := newLogger(logLevel)
	ctrl.SetLogger(logger)

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         leaderElect,
		LeaderElectionID:       "backrest-volsync-operator.backrest.garethgeorge.com",
		LeaseDuration:          ptr(30 * time.Second),
		RenewDeadline:          ptr(20 * time.Second),
		RetryPeriod:            ptr(5 * time.Second),
	})
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&controllers.BackrestVolSyncBindingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	logger.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func ptr[T any](v T) *T { return &v }

func newLogger(level string) logr.Logger {
	// stdr is simple and dependency-light; controller-runtime will add context fields.
	_ = level
	return stdr.New(nil)
}
