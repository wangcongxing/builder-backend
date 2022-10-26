// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongodb

import (
	"context"

	"github.com/illa-family/builder-backend/pkg/plugins/common"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Connector struct {
	Resource Options
	Action   Query
}

func (m *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format mongodb simple options
	if err := mapstructure.Decode(resourceOptions, &m.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate simple options
	validate := validator.New()
	if err := validate.Struct(m.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate specific options
	if m.Resource.ConfigType == GUI_OPTIONS {
		var mOptions GUIOptions
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
		if err := validate.Struct(mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
	} else if m.Resource.ConfigType == URI_OPTIONS {
		var mOptions URIOptions
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
		if err := validate.Struct(mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
	}

	return common.ValidateResult{Valid: true}, nil
}

func (m *Connector) ValidateActionOptions(actionOptions map[string]interface{}) (common.ValidateResult, error) {

	return common.ValidateResult{Valid: true}, nil
}

func (m *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get mongodb connection
	client, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer client.Disconnect(context.TODO())

	// test mongodb connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	return common.ConnectionResult{Success: true}, nil
}

func (m *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {

	return common.MetaInfoResult{
		Success: true,
		Schema:  nil,
	}, nil
}

func (m *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}) (common.RuntimeResult, error) {

	return common.RuntimeResult{Success: false}, nil
}
