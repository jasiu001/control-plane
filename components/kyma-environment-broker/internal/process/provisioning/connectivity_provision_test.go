package provisioning

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Peripli/service-manager-cli/pkg/types"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/logger"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/servicemanager"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestConnectivityProvisioningStep_Run(t *testing.T) {
	// given
	repo := storage.NewMemoryStorage().Operations()
	clientFactory := servicemanager.NewFakeServiceManagerClientFactory([]types.ServiceOffering{}, []types.ServicePlan{})
	clientFactory.SynchronousProvisioning()
	operation := internal.ProvisioningOperation{
		Operation: internal.Operation{
			InstanceDetails: internal.InstanceDetails{
				Connectivity: internal.ConnectivityData{Instance: internal.ServiceManagerInstanceInfo{
					BrokerID:  "broker-id",
					ServiceID: "svc-id",
					PlanID:    "plan-id",
				}},
				ShootDomain: "conn-test.sap.com",
			},
		},
		SMClientFactory: clientFactory,
	}
	offeringStep := NewServiceManagerOfferingStep("Connectivity_Offering",
		ConnectivityOfferingName, ConnectivityPlanName, func(op *internal.ProvisioningOperation) *internal.ServiceManagerInstanceInfo {
			return &op.Connectivity.Instance
		}, repo)

	provisionStep := NewConnectivityProvisionStep(repo)
	repo.InsertProvisioningOperation(operation)

	log := logger.NewLogDummy()
	// when
	operation, retry, err := offeringStep.Run(operation, log)
	require.NoError(t, err)
	require.Zero(t, retry)

	operation, retry, err = provisionStep.Run(operation, logger.NewLogDummy())

	// then
	assert.NoError(t, err)
	assert.Zero(t, retry)
	assert.NotEmpty(t, operation.Connectivity.Instance.InstanceID)
	assert.False(t, operation.Connectivity.Instance.Provisioned)
	assert.True(t, operation.Connectivity.Instance.ProvisioningTriggered)
	clientFactory.AssertProvisionCalled(t, servicemanager.InstanceKey{
		BrokerID:   "broker-id",
		InstanceID: operation.Connectivity.Instance.InstanceID,
		ServiceID:  "svc-id",
		PlanID:     "plan-id",
	})
}
