package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

type LmsTenantProvider interface {
	ProvideLMSTenantID(name, region string) (string, error)
}

// provideLmsTenantStep creates (if not exists) LMS tenant and provides its ID.
// The step does not breaks the provisioning flow.
type provideLmsTenantStep struct {
	tenantProvider   LmsTenantProvider
	operationManager *process.ProvisionOperationManager
	regionOverride   string
}

func NewProvideLmsTenantStep(tp LmsTenantProvider, repo storage.Operations, regionOverride string) *provideLmsTenantStep {
	return &provideLmsTenantStep{
		tenantProvider:   tp,
		operationManager: process.NewProvisionOperationManager(repo),
		regionOverride:   regionOverride,
	}
}

func (s *provideLmsTenantStep) Name() string {
	return "Create_LMS_Tenant"
}

func (s *provideLmsTenantStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if operation.Lms.TenantID != "" {
		return operation, 0, nil
	}

	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		msg := fmt.Sprintf("Unable to get provisioning parameters: %s", err.Error())
		logger.Errorf(msg)
		return s.operationManager.OperationFailed(operation, msg)
	}
	region := s.provideRegion(pp.Parameters.Region)

	lmsTenantID, err := s.tenantProvider.ProvideLMSTenantID(pp.ErsContext.GlobalAccountID, region)
	if err != nil {
		logger.Warnf("Unable to get tenant for GlobalaccountID/region %s/%s: %s", pp.ErsContext.GlobalAccountID, region, err.Error())
		since := time.Since(operation.UpdatedAt)
		if since < 3*time.Minute {
			return operation, 30 * time.Second, nil
		}

		logger.Errorf("Unable to get tenant, setting LMS failed")
		// if it is not possible to request tenant - set LMS failed and process next steps
		operation.Lms.Failed = true
		modifiedOp, repeat := s.operationManager.UpdateOperation(operation)
		if repeat != 0 {
			logger.Errorf("cannot save operation")
			return operation, time.Second, nil
		}
		return modifiedOp, 0, nil
	}

	operation.Lms.TenantID = lmsTenantID
	if operation.Lms.RequestedAt.IsZero() {
		operation.Lms.RequestedAt = time.Now()
	}

	op, repeat := s.operationManager.UpdateOperation(operation)
	if repeat != 0 {
		logger.Errorf("cannot save LMS tenant ID")
		return operation, time.Second, nil
	}

	return op, 0, nil
}

var lmsRegionsMap = map[string]string{
	"westeurope":    "eu",
	"eastus":        "us",
	"eastus2":       "us",
	"centralus":     "us",
	"northeurope":   "eu",
	"southeastasia": "aus",
	"japaneast":     "aus",
	"westus2":       "eu",
	"uksouth":       "eu",
	"FranceCentral": "eu",
	"EastUS2EUAP":   "us",
	"uaenorth":      "eu",
}

func (s *provideLmsTenantStep) provideRegion(r *string) string {
	if s.regionOverride != "" {
		return s.regionOverride
	}
	if r == nil {
		return "eu"
	}
	region, found := lmsRegionsMap[*r]
	if !found {
		return "eu"
	}
	return region
}
