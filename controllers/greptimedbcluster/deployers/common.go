package deployers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/pkg/deployer"
)

const (
	GreptimeComponentName = "app.greptime.io/component"
)

func UpdateStatus(ctx context.Context, input *v1alpha1.GreptimeDBCluster, kc client.Client, opts ...client.UpdateOption) error {
	cluster := input.DeepCopy()
	status := cluster.Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		objectKey := client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name}
		if err = kc.Get(ctx, objectKey, cluster); err != nil {
			return
		}
		cluster.Status = status
		return kc.Status().Update(ctx, cluster, opts...)
	})
}

// CommonDeployer is the common deployer for all components of GreptimeDBCluster.
type CommonDeployer struct {
	client.Client
	Scheme *runtime.Scheme

	deployer.DefaultDeployer
}

func NewFromManager(mgr ctrl.Manager) *CommonDeployer {
	return &CommonDeployer{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),

		DefaultDeployer: deployer.DefaultDeployer{
			Client: mgr.GetClient(),
		},
	}
}

func (c *CommonDeployer) ResourceName(clusterName string, componentKind v1alpha1.ComponentKind) string {
	return clusterName + "-" + string(componentKind)
}

func (c *CommonDeployer) GetCluster(crdObject client.Object) (*v1alpha1.GreptimeDBCluster, error) {
	cluster, ok := crdObject.(*v1alpha1.GreptimeDBCluster)
	if !ok {
		return nil, fmt.Errorf("the object is not GreptimeDBCluster")
	}
	return cluster, nil
}
