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

package v4

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
)

type RuleService struct {
	//
}

func (s *RuleService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices/:serviceId/rules", Func: s.AddRule},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/rules", Func: s.GetRules},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/rules/:rule_id", Func: s.UpdateRule},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId/rules/:rule_id", Func: s.DeleteRule},
	}
}
func (s *RuleService) AddRule(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	rule := map[string][]*pb.AddOrUpdateServiceRule{}
	err = json.Unmarshal(message, &rule)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	resp, err := core.ServiceAPI.AddRule(r.Context(), &pb.AddServiceRulesRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
		Rules:     rule["rules"],
	})
	if err != nil {
		log.Errorf(err, "add rule failed")
		rest.WriteError(w, pb.ErrInternal, "add rule failed")
		return
	}
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *RuleService) DeleteRule(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get(":rule_id")
	ids := strings.Split(id, ",")

	resp, _ := core.ServiceAPI.DeleteRule(r.Context(), &pb.DeleteServiceRulesRequest{
		ServiceId: query.Get(":serviceId"),
		RuleIds:   ids,
	})
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *RuleService) UpdateRule(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	rule := pb.AddOrUpdateServiceRule{}
	err = json.Unmarshal(message, &rule)
	if err != nil {
		log.Errorf(err, "invalid json: %s", util.BytesToStringWithNoCopy(message))
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	query := r.URL.Query()
	resp, err := core.ServiceAPI.UpdateRule(r.Context(), &pb.UpdateServiceRuleRequest{
		ServiceId: query.Get(":serviceId"),
		RuleId:    query.Get(":rule_id"),
		Rule:      &rule,
	})
	if err != nil {
		log.Errorf(err, "update rule failed")
		rest.WriteError(w, pb.ErrInternal, "update rule failed")
		return
	}
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *RuleService) GetRules(w http.ResponseWriter, r *http.Request) {
	resp, _ := core.ServiceAPI.GetRule(r.Context(), &pb.GetServiceRulesRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	})
	rest.WriteResponse(w, r, resp.Response, resp)
}
