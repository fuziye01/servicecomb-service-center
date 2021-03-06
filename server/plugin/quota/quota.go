/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package quota

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/metrics"
)

const QUOTA plugin.Kind = "quota"

const (
	defaultServiceLimit  = 50000
	defaultInstanceLimit = 150000
	defaultSchemaLimit   = 100
	defaultRuleLimit     = 100
	defaultTagLimit      = 100
	defaultAccountLimit  = 1000
	defaultRoleLimit     = 100
)

const (
	TypeRule ResourceType = iota
	TypeSchema
	TypeTag
	TypeService
	TypeInstance
	TypeAccount
	TypeRole
)

var (
	DefaultServiceQuota  = defaultServiceLimit
	DefaultInstanceQuota = defaultInstanceLimit
	DefaultSchemaQuota   = defaultSchemaLimit
	DefaultTagQuota      = defaultTagLimit
	DefaultRuleQuota     = defaultRuleLimit
	DefaultAccountQuota  = defaultAccountLimit
	DefaultRoleQuota     = defaultRoleLimit
)

func Init() {
	DefaultServiceQuota = config.GetInt("quota.cap.service.limit", defaultServiceLimit, config.WithENV("QUOTA_SERVICE"))
	DefaultInstanceQuota = config.GetInt("quota.cap.instance.limit", defaultInstanceLimit, config.WithENV("QUOTA_INSTANCE"))
	DefaultSchemaQuota = config.GetInt("quota.cap.schema.limit", defaultSchemaLimit, config.WithENV("QUOTA_SCHEMA"))
	DefaultTagQuota = config.GetInt("quota.cap.tag.limit", defaultTagLimit, config.WithENV("QUOTA_TAG"))
	DefaultRuleQuota = config.GetInt("quota.cap.rule.limit", defaultRuleLimit, config.WithENV("QUOTA_RULE"))
	DefaultAccountQuota = config.GetInt("quota.cap.account.limit", defaultAccountLimit, config.WithENV("QUOTA_ACCOUNT"))
	DefaultRoleQuota = config.GetInt("quota.cap.role.limit", defaultRoleLimit, config.WithENV("QUOTA_ROLE"))
}

type ApplyQuotaResource struct {
	QuotaType     ResourceType
	DomainProject string
	ServiceID     string
	QuotaSize     int64
}

func NewApplyQuotaResource(quotaType ResourceType, domainProject, serviceID string, quotaSize int64) *ApplyQuotaResource {
	return &ApplyQuotaResource{
		quotaType,
		domainProject,
		serviceID,
		quotaSize,
	}
}

type Manager interface {
	RemandQuotas(ctx context.Context, quotaType ResourceType)
	GetQuota(ctx context.Context, t ResourceType) int64
}

type ResourceType int

func (r ResourceType) String() string {
	switch r {
	case TypeRule:
		return "RULE"
	case TypeSchema:
		return "SCHEMA"
	case TypeTag:
		return "TAG"
	case TypeService:
		return "SERVICE"
	case TypeInstance:
		return "INSTANCE"
	case TypeAccount:
		return "ACCOUNT"
	case TypeRole:
		return "ROLE"
	default:
		return "RESOURCE" + strconv.Itoa(int(r))
	}
}

//申请配额sourceType serviceinstance servicetype
func Apply(ctx context.Context, res *ApplyQuotaResource) *errsvc.Error {
	if res == nil {
		err := errors.New("invalid parameters")
		log.Errorf(err, "quota check failed")
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	limitQuota := plugin.Plugins().Instance(QUOTA).(Manager).GetQuota(ctx, res.QuotaType)
	curNum, err := GetResourceUsage(ctx, res)
	if err != nil {
		log.Errorf(err, "%s quota check failed", res.QuotaType)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if curNum+res.QuotaSize > limitQuota {
		mes := fmt.Sprintf("no quota to create %s, max num is %d, curNum is %d, apply num is %d",
			res.QuotaType, limitQuota, curNum, res.QuotaSize)
		log.Errorf(nil, mes)
		return pb.NewError(pb.ErrNotEnoughQuota, mes)
	}
	return nil
}

func Remand(ctx context.Context, quotaType ResourceType) {
	plugin.Plugins().Instance(QUOTA).(Manager).RemandQuotas(ctx, quotaType)
}
func GetResourceUsage(ctx context.Context, res *ApplyQuotaResource) (int64, error) {
	serviceID := res.ServiceID
	switch res.QuotaType {
	case TypeService:
		return metrics.GetTotalService(util.ParseDomain(ctx)), nil
	case TypeInstance:
		usage := metrics.GetTotalInstance(util.ParseDomain(ctx))
		return usage, nil
	case TypeRule:
		{
			resp, err := datasource.GetMetadataManager().GetRules(ctx, &pb.GetServiceRulesRequest{
				ServiceId: serviceID,
			})
			if err != nil {
				return 0, err
			}
			return int64(len(resp.Rules)), nil
		}
	case TypeSchema:
		{
			resp, err := datasource.GetMetadataManager().GetAllSchemas(ctx, &pb.GetAllSchemaRequest{
				ServiceId:  serviceID,
				WithSchema: false,
			})
			if err != nil {
				return 0, err
			}
			return int64(len(resp.Schemas)), nil
		}
	case TypeTag:
		// always re-create the service old tags
		return 0, nil
	case TypeRole:
		{
			_, used, err := datasource.GetRoleManager().ListRole(ctx)
			if err != nil {
				return 0, err
			}
			return used, nil
		}
	case TypeAccount:
		{
			_, used, err := datasource.GetAccountManager().ListAccount(ctx)
			if err != nil {
				return 0, err
			}
			return used, nil
		}
	default:
		return 0, fmt.Errorf("not define quota type '%s'", res.QuotaType)
	}
}
