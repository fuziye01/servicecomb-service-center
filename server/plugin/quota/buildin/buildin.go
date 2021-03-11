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

package buildin

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{Kind: quota.QUOTA, Name: "buildin", New: New})
}

func New() mgr.Instance {
	quota.Init()
	log.Infof("quota init, service: %d, instance: %d, schema: %d/service, tag: %d/service, rule: %d/service",
		quota.DefaultServiceQuota, quota.DefaultInstanceQuota,
		quota.DefaultSchemaQuota, quota.DefaultTagQuota, quota.DefaultRuleQuota)
	return &Quota{}
}

type Quota struct {
}

func (q *Quota) GetQuota(ctx context.Context, t quota.ResourceType) int64 {
	switch t {
	case quota.TypeInstance:
		return int64(quota.DefaultInstanceQuota)
	case quota.TypeService:
		return int64(quota.DefaultServiceQuota)
	case quota.TypeRule:
		return int64(quota.DefaultRuleQuota)
	case quota.TypeSchema:
		return int64(quota.DefaultSchemaQuota)
	case quota.TypeTag:
		return int64(quota.DefaultTagQuota)
	default:
		return 0
	}
}

//向配额中心上报配额使用量
func (q *Quota) RemandQuotas(ctx context.Context, quotaType quota.ResourceType) {
	df, ok := mgr.DynamicPluginFunc(quota.QUOTA, "RemandQuotas").(func(context.Context, quota.ResourceType))
	if ok {
		df(ctx, quotaType)
		return
	}
}
