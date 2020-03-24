// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2019 HAProxy Technologies
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
//

package discovery

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"

	"github.com/haproxytech/models"
)

// GetClusterOKCode is the HTTP code returned for type GetClusterOK
const GetClusterOKCode int = 200

/*GetClusterOK Success

swagger:response getClusterOK
*/
type GetClusterOK struct {

	/*
	  In: Body
	*/
	Payload *models.ClusterSettings `json:"body,omitempty"`
}

// NewGetClusterOK creates GetClusterOK with default headers values
func NewGetClusterOK() *GetClusterOK {

	return &GetClusterOK{}
}

// WithPayload adds the payload to the get cluster o k response
func (o *GetClusterOK) WithPayload(payload *models.ClusterSettings) *GetClusterOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get cluster o k response
func (o *GetClusterOK) SetPayload(payload *models.ClusterSettings) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetClusterOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetClusterDefault General Error

swagger:response getClusterDefault
*/
type GetClusterDefault struct {
	_statusCode int
	/*Configuration file version

	 */
	ConfigurationVersion int64 `json:"Configuration-Version"`

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetClusterDefault creates GetClusterDefault with default headers values
func NewGetClusterDefault(code int) *GetClusterDefault {
	if code <= 0 {
		code = 500
	}

	return &GetClusterDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get cluster default response
func (o *GetClusterDefault) WithStatusCode(code int) *GetClusterDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get cluster default response
func (o *GetClusterDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithConfigurationVersion adds the configurationVersion to the get cluster default response
func (o *GetClusterDefault) WithConfigurationVersion(configurationVersion int64) *GetClusterDefault {
	o.ConfigurationVersion = configurationVersion
	return o
}

// SetConfigurationVersion sets the configurationVersion to the get cluster default response
func (o *GetClusterDefault) SetConfigurationVersion(configurationVersion int64) {
	o.ConfigurationVersion = configurationVersion
}

// WithPayload adds the payload to the get cluster default response
func (o *GetClusterDefault) WithPayload(payload *models.Error) *GetClusterDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get cluster default response
func (o *GetClusterDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetClusterDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	// response header Configuration-Version

	configurationVersion := swag.FormatInt64(o.ConfigurationVersion)
	if configurationVersion != "" {
		rw.Header().Set("Configuration-Version", configurationVersion)
	}

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
