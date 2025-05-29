// Copyright 2023 Greptime Team
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

package dbconfig

import (
	"fmt"

	"k8s.io/utils/ptr"

	"github.com/GreptimeTeam/greptimedb-operator/apis/v1alpha1"
	k8sutil "github.com/GreptimeTeam/greptimedb-operator/pkg/util/k8s"
)

var _ Config = &MetaConfig{}

// MetaConfig is the configuration for the meta.
type MetaConfig struct {
	// Enable region failover.
	EnableRegionFailover *bool `tomlmapping:"enable_region_failover"`

	// If it's not empty, the meta will store all data with this key prefix.
	StoreKeyPrefix *string `tomlmapping:"store_key_prefix"`

	// The store addrs.
	StoreAddrs []string `tomlmapping:"store_addrs"`

	// The meta table name.
	MetaTableName *string `tomlmapping:"meta_table_name"`

	// The backend storage type.
	Backend *string `tomlmapping:"backend"`

	// The wal provider.
	WalProvider *string `tomlmapping:"wal.provider"`

	// The kafka broker endpoints.
	WalBrokerEndpoints []string `tomlmapping:"wal.broker_endpoints"`

	// LoggingConfig is the configuration for the logging.
	LoggingConfig `tomlmapping:",inline"`

	// InputConfig is from config field of cluster spec.
	InputConfig string
}

// ConfigureByCluster configures the meta config by the given cluster.
func (c *MetaConfig) ConfigureByCluster(cluster *v1alpha1.GreptimeDBCluster, roleSpec v1alpha1.RoleSpec) error {
	if roleSpec.GetRoleKind() != v1alpha1.MetaRoleKind {
		return fmt.Errorf("invalid role kind: %s", roleSpec.GetRoleKind())
	}

	metaSpec, ok := roleSpec.(*v1alpha1.MetaSpec)
	if !ok {
		return fmt.Errorf("invalid role spec type: %T", roleSpec)
	}

	c.EnableRegionFailover = ptr.To(metaSpec.IsEnableRegionFailover())

	if prefix := metaSpec.GetStoreKeyPrefix(); prefix != "" {
		c.StoreKeyPrefix = ptr.To(prefix)
	}

	if err := c.configureBackendStorage(metaSpec, cluster.GetNamespace()); err != nil {
		return err
	}

	if cfg := metaSpec.GetConfig(); cfg != "" {
		if err := c.SetInputConfig(cfg); err != nil {
			return err
		}
	}

	if kafka := cluster.GetWALProvider().GetKafkaWAL(); kafka != nil {
		c.WalProvider = ptr.To("kafka")
		c.WalBrokerEndpoints = kafka.GetBrokerEndpoints()
	}

	c.ConfigureLogging(metaSpec.GetLogging())

	return nil
}

// ConfigureByStandalone is not need to implement in cluster mode.
func (c *MetaConfig) ConfigureByStandalone(_ *v1alpha1.GreptimeDBStandalone) error {
	return nil
}

// Kind returns the component kind of the meta.
func (c *MetaConfig) Kind() v1alpha1.RoleKind {
	return v1alpha1.MetaRoleKind
}

// GetInputConfig returns the input config of the meta.
func (c *MetaConfig) GetInputConfig() string {
	return c.InputConfig
}

// SetInputConfig sets the input config of the meta.
func (c *MetaConfig) SetInputConfig(inputConfig string) error {
	c.InputConfig = inputConfig
	return nil
}

func (c *MetaConfig) configureBackendStorage(spec *v1alpha1.MetaSpec, namespace string) error {
	if etcd := spec.GetBackendStorage().GetEtcdStorage(); etcd != nil {
		c.Backend = ptr.To("etcd_store")
		c.StoreAddrs = etcd.GetEndpoints()
		if prefix := etcd.GetStoreKeyPrefix(); prefix != "" {
			c.StoreKeyPrefix = ptr.To(prefix)
		}
	}

	if mysql := spec.GetBackendStorage().GetMySQLStorage(); mysql != nil {
		c.Backend = ptr.To("mysql_store")
		conn, err := c.generateMySQLConnectionString(mysql, namespace)
		if err != nil {
			return err
		}
		c.StoreAddrs = []string{conn}
		c.MetaTableName = ptr.To(mysql.Table)
	}

	if postgresql := spec.GetBackendStorage().GetPostgreSQLStorage(); postgresql != nil {
		c.Backend = ptr.To("postgres_store")
		conn, err := c.generatePostgreSQLConnectionString(postgresql, namespace)
		if err != nil {
			return err
		}
		c.StoreAddrs = []string{conn}
	}

	// Compatibility with the old api version.
	if len(spec.EtcdEndpoints) > 0 {
		c.Backend = ptr.To("etcd_store")
		c.StoreAddrs = spec.EtcdEndpoints
		if prefix := spec.GetStoreKeyPrefix(); prefix != "" {
			c.StoreKeyPrefix = ptr.To(prefix)
		}
	}

	return nil
}

// generateMySQLConnectionString generates the connection string for the MySQL database.
// For example, the connection string looks like:
//
//	mysql://root:greptimedb-meta@mysql.default.svc.cluster.local:3306/metasrv
func (c *MetaConfig) generateMySQLConnectionString(mysql *v1alpha1.MySQLStorage, namespace string) (string, error) {
	data, err := k8sutil.GetSecretsData(namespace, mysql.CredentialsSecretName, []string{v1alpha1.MetaDatabaseUsernameKey, v1alpha1.MetaDatabasePasswordKey})
	if err != nil {
		return "", err
	}
	username := string(data[0])
	password := string(data[1])

	conn := fmt.Sprintf("mysql://%s:%s@%s:%d/%s", username, password, mysql.Host, mysql.Port, mysql.Database)

	return conn, nil
}

// generatePostgreSQLConnectionString generates the connection string for the PostgreSQL database.
// For example, the connection string looks like:
//
//	postgres://root:greptimedb-meta@postgresql.default.svc.cluster.local:5432/metasrv
func (c *MetaConfig) generatePostgreSQLConnectionString(postgresql *v1alpha1.PostgreSQLStorage, namespace string) (string, error) {
	data, err := k8sutil.GetSecretsData(namespace, postgresql.CredentialsSecretName, []string{v1alpha1.MetaDatabaseUsernameKey, v1alpha1.MetaDatabasePasswordKey})
	if err != nil {
		return "", err
	}
	username := string(data[0])
	password := string(data[1])

	conn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, postgresql.Host, postgresql.Port, postgresql.Database)

	return conn, nil
}
