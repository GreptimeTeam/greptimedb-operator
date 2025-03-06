// Copyright 2024 Greptime Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	greptimev1alpha1 "github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	"github.com/GreptimeTeam/greptimedb-operator/controllers/common"
)

const (
	// MaxRetryTimesForAvailablePort is the maximum retry times to get an available random port for port forwarding.
	MaxRetryTimesForAvailablePort = 100

	// MinPort is the minimum available port for port forwarding.
	MinPort = 10000

	// MaxPort is the maximum available port for port forwarding.
	MaxPort = 30000

	// DefaultTimeout is the default timeout for the e2e tests.
	DefaultTimeout = 1 * time.Minute

	// DefaultEtcdNamespace is the default namespace for the etcd cluster.
	DefaultEtcdNamespace = "etcd-cluster"

	// DefaultEtcdPodName is the default pod name for the etcd cluster.
	DefaultEtcdPodName = "etcd-0"
)

// Helper is a helper utility for e2e tests.
type Helper struct {
	client.Client
}

// New creates a new Helper instance.
func New(client client.Client) *Helper {
	return &Helper{
		Client: client,
	}
}

// RunSQLTest runs the SQL test of the given SQL file on the given address by using the pgx driver.
// TODO(zyy17): We need to use sqlness in the future.
func (h *Helper) RunSQLTest(ctx context.Context, addr string, sqlFile string) error {
	// connect public database by default.
	url := fmt.Sprintf("postgres://postgres@%s/public?sslmode=disable", addr)

	fmt.Printf("Connecting to %s\n", url)
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	data, err := os.ReadFile(sqlFile)
	if err != nil {
		return err
	}

	_, err = conn.Exec(context.Background(), string(data))
	if err != nil {
		return err
	}

	return nil
}

// RunHTTPTest runs the HTTP request to the specified hostname with the given data.
func (h *Helper) RunHTTPTest(hostname string, data string) error {
	req, err := http.NewRequest("POST", hostname, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to run HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP response status: got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	return nil
}

// AddIPToHosts adds the loadBalancer IP and hostname to the /etc/hosts file.
func (h *Helper) AddIPToHosts(ip string, hostname string) error {
	const (
		hostsFile = "/etc/hosts"
	)

	// Read the current contents of the /etc/hosts file
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check if the entry already exists
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, hostname) && strings.Contains(line, ip) {
			return nil
		}
	}

	// Prepare the new entry
	newEntry := fmt.Sprintf("%s\t%s\n", ip, hostname)

	if err := os.Chmod(hostsFile, 0666); err != nil {
		return fmt.Errorf("failed to change permissions on %s: %w", hostsFile, err)
	}

	// Append the new entry to the /etc/hosts file
	f, err := os.OpenFile(hostsFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open /etc/hosts for writing: %w", err)
	}
	defer f.Close()

	if _, err = f.WriteString(newEntry); err != nil {
		return fmt.Errorf("failed to write ingress ip to /etc/hosts: %w", err)
	}

	return nil
}

// GetIngressIP returns the ingress IP of the given service. It is assumed that the service is of type LoadBalancer.
func (h *Helper) GetIngressIP(ctx context.Context, namespace, name string) (string, error) {
	var svc corev1.Service
	if err := h.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &svc); err != nil {
		return "", err
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("no ingress found for service %s", name)
	}

	return svc.Status.LoadBalancer.Ingress[0].IP, nil
}

// GetPhase returns the phase of GreptimeDBCluster or GreptimeDBStandalone object.
func (h *Helper) GetPhase(ctx context.Context, namespace, name string, object client.Object) (greptimev1alpha1.Phase, error) {
	if err := h.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, object); err != nil {
		return "", err
	}

	// Check the type of the object and return the phase accordingly.
	switch o := object.(type) {
	case *greptimev1alpha1.GreptimeDBCluster:
		return o.Status.ClusterPhase, nil
	case *greptimev1alpha1.GreptimeDBStandalone:
		return o.Status.StandalonePhase, nil
	default:
		return "", fmt.Errorf("unknown object type")
	}
}

// GetPVCs returns the PVC list of the given component.
func (h *Helper) GetPVCs(ctx context.Context, namespace, name string, kind greptimev1alpha1.ComponentKind, fsType common.FileStorageType) ([]corev1.PersistentVolumeClaim, error) {
	return common.GetPVCs(ctx, h.Client, namespace, name, kind, fsType)
}

// CleanEtcdData cleans up all data in etcd by executing the etcdctl command in the given pod.
func (h *Helper) CleanEtcdData(ctx context.Context) error {
	var (
		cli  = "kubectl"
		args = []string{"exec", "-n", DefaultEtcdNamespace, DefaultEtcdPodName, "--", "etcdctl", "del", "--prefix", ""}
	)

	cmd := exec.CommandContext(ctx, cli, args...)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// LoadCR loads the CR from the input yaml file and unmarshals it into the given object.
func (h *Helper) LoadCR(inputFile string, dstObject client.Object) error {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, dstObject); err != nil {
		return err
	}

	return nil
}

// PortForward uses kubectl to port forward the service to a local random available port.
func (h *Helper) PortForward(ctx context.Context, namespace, svc string, srcPort int) (string, error) {
	randomPort, err := h.getRandomAvailablePort(MinPort, MaxPort, MaxRetryTimesForAvailablePort)
	if err != nil {
		return "", err
	}

	// Start the port forwarding in a goroutine.
	go func() {
		var (
			cli  = "kubectl"
			args = []string{"port-forward", "-n", namespace, fmt.Sprintf("svc/%s", svc), fmt.Sprintf("%d:%d", randomPort, srcPort)}
		)

		fmt.Printf("Use kubectl to port forwarding %s/%s:%d to %d\n", namespace, svc, srcPort, randomPort)
		cmd := exec.CommandContext(ctx, cli, args...)
		if err := cmd.Run(); err != nil {
			return
		}
	}()

	return fmt.Sprintf("localhost:%d", randomPort), nil
}

// KillPortForwardProcess kills the port forwarding process.
func (h *Helper) KillPortForwardProcess() {
	cmd := exec.Command("pkill", "-f", "kubectl port-forward")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

// TimestampInMillisecond returns the current timestamp in milliseconds.
func (h *Helper) TimestampInMillisecond() int64 {
	now := time.Now()
	return now.Unix()*1000 + int64(now.Nanosecond())/1e6
}

// getRandomAvailablePort returns a random available port in the [minPort, maxPort] range.
func (h *Helper) getRandomAvailablePort(minPort, maxPort, retryTimes int) (int, error) {
	// Try 100 times to get a random available port.
	for i := 0; i < retryTimes; i++ {
		port := rand.Intn(maxPort-minPort+1) + minPort
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		listener.Close()
		return port, nil
	}

	return 0, fmt.Errorf("failed to get a random available port")
}
