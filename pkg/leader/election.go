package leader

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// LeaderElection handles leader election for high availability
type LeaderElection struct {
	clientset kubernetes.Interface
	namespace string
	name      string
	identity  string
}

// NewLeaderElection creates a new leader election instance
func NewLeaderElection(clientset kubernetes.Interface, namespace, name string) *LeaderElection {
	hostname, _ := os.Hostname()
	identity := fmt.Sprintf("%s-%d", hostname, os.Getpid())

	return &LeaderElection{
		clientset: clientset,
		namespace: namespace,
		name:      name,
		identity:  identity,
	}
}

// Run starts leader election process
func (le *LeaderElection) Run(ctx context.Context, onStartedLeading, onStoppedLeading func(context.Context)) error {
	// Create lock
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      le.name,
			Namespace: le.namespace,
		},
		Client: le.clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: le.identity,
		},
	}

	// Leader election config
	config := leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Info().Str("identity", le.identity).Msg("Started leading")
				if onStartedLeading != nil {
					onStartedLeading(ctx)
				}
			},
			OnStoppedLeading: func() {
				log.Info().Str("identity", le.identity).Msg("Stopped leading")
				if onStoppedLeading != nil {
					onStoppedLeading(ctx)
				}
			},
			OnNewLeader: func(identity string) {
				if identity == le.identity {
					return
				}
				log.Info().Str("leader", identity).Msg("New leader elected")
			},
		},
	}

	// Start leader election
	log.Info().Str("identity", le.identity).Str("name", le.name).Msg("Starting leader election")
	leaderelection.RunOrDie(ctx, config)
	
	return nil
}
