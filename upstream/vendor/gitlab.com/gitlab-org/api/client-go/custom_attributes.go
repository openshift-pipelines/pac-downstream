//
// Copyright 2021, Sander van Harmelen
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

package gitlab

import (
	"fmt"
	"net/http"
)

type (
	CustomAttributesServiceInterface interface {
		ListCustomUserAttributes(user int, options ...RequestOptionFunc) ([]*CustomAttribute, *Response, error)
		ListCustomGroupAttributes(group int, options ...RequestOptionFunc) ([]*CustomAttribute, *Response, error)
		ListCustomProjectAttributes(project int, options ...RequestOptionFunc) ([]*CustomAttribute, *Response, error)
		GetCustomUserAttribute(user int, key string, options ...RequestOptionFunc) (*CustomAttribute, *Response, error)
		GetCustomGroupAttribute(group int, key string, options ...RequestOptionFunc) (*CustomAttribute, *Response, error)
		GetCustomProjectAttribute(project int, key string, options ...RequestOptionFunc) (*CustomAttribute, *Response, error)
		SetCustomUserAttribute(user int, c CustomAttribute, options ...RequestOptionFunc) (*CustomAttribute, *Response, error)
		SetCustomGroupAttribute(group int, c CustomAttribute, options ...RequestOptionFunc) (*CustomAttribute, *Response, error)
		SetCustomProjectAttribute(project int, c CustomAttribute, options ...RequestOptionFunc) (*CustomAttribute, *Response, error)
		DeleteCustomUserAttribute(user int, key string, options ...RequestOptionFunc) (*Response, error)
		DeleteCustomGroupAttribute(group int, key string, options ...RequestOptionFunc) (*Response, error)
		DeleteCustomProjectAttribute(project int, key string, options ...RequestOptionFunc) (*Response, error)
	}

	// CustomAttributesService handles communication with the group, project and
	// user custom attributes related methods of the GitLab API.
	//
	// GitLab API docs: https://docs.gitlab.com/api/custom_attributes/
	CustomAttributesService struct {
		client *Client
	}
)

var _ CustomAttributesServiceInterface = (*CustomAttributesService)(nil)

// CustomAttribute struct is used to unmarshal response to api calls.
//
// GitLab API docs: https://docs.gitlab.com/api/custom_attributes/
type CustomAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ListCustomUserAttributes lists the custom attributes of the specified user.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#list-custom-attributes
func (s *CustomAttributesService) ListCustomUserAttributes(user int, options ...RequestOptionFunc) ([]*CustomAttribute, *Response, error) {
	return s.listCustomAttributes("users", user, options...)
}

// ListCustomGroupAttributes lists the custom attributes of the specified group.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#list-custom-attributes
func (s *CustomAttributesService) ListCustomGroupAttributes(group int, options ...RequestOptionFunc) ([]*CustomAttribute, *Response, error) {
	return s.listCustomAttributes("groups", group, options...)
}

// ListCustomProjectAttributes lists the custom attributes of the specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#list-custom-attributes
func (s *CustomAttributesService) ListCustomProjectAttributes(project int, options ...RequestOptionFunc) ([]*CustomAttribute, *Response, error) {
	return s.listCustomAttributes("projects", project, options...)
}

func (s *CustomAttributesService) listCustomAttributes(resource string, id int, options ...RequestOptionFunc) ([]*CustomAttribute, *Response, error) {
	u := fmt.Sprintf("%s/%d/custom_attributes", resource, id)
	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var cas []*CustomAttribute
	resp, err := s.client.Do(req, &cas)
	if err != nil {
		return nil, resp, err
	}
	return cas, resp, nil
}

// GetCustomUserAttribute returns the user attribute with a speciifc key.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#single-custom-attribute
func (s *CustomAttributesService) GetCustomUserAttribute(user int, key string, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	return s.getCustomAttribute("users", user, key, options...)
}

// GetCustomGroupAttribute returns the group attribute with a speciifc key.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#single-custom-attribute
func (s *CustomAttributesService) GetCustomGroupAttribute(group int, key string, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	return s.getCustomAttribute("groups", group, key, options...)
}

// GetCustomProjectAttribute returns the project attribute with a speciifc key.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#single-custom-attribute
func (s *CustomAttributesService) GetCustomProjectAttribute(project int, key string, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	return s.getCustomAttribute("projects", project, key, options...)
}

func (s *CustomAttributesService) getCustomAttribute(resource string, id int, key string, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	u := fmt.Sprintf("%s/%d/custom_attributes/%s", resource, id, key)
	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var ca *CustomAttribute
	resp, err := s.client.Do(req, &ca)
	if err != nil {
		return nil, resp, err
	}
	return ca, resp, nil
}

// SetCustomUserAttribute sets the custom attributes of the specified user.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#set-custom-attribute
func (s *CustomAttributesService) SetCustomUserAttribute(user int, c CustomAttribute, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	return s.setCustomAttribute("users", user, c, options...)
}

// SetCustomGroupAttribute sets the custom attributes of the specified group.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#set-custom-attribute
func (s *CustomAttributesService) SetCustomGroupAttribute(group int, c CustomAttribute, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	return s.setCustomAttribute("groups", group, c, options...)
}

// SetCustomProjectAttribute sets the custom attributes of the specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#set-custom-attribute
func (s *CustomAttributesService) SetCustomProjectAttribute(project int, c CustomAttribute, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	return s.setCustomAttribute("projects", project, c, options...)
}

func (s *CustomAttributesService) setCustomAttribute(resource string, id int, c CustomAttribute, options ...RequestOptionFunc) (*CustomAttribute, *Response, error) {
	u := fmt.Sprintf("%s/%d/custom_attributes/%s", resource, id, c.Key)
	req, err := s.client.NewRequest(http.MethodPut, u, c, options)
	if err != nil {
		return nil, nil, err
	}

	ca := new(CustomAttribute)
	resp, err := s.client.Do(req, ca)
	if err != nil {
		return nil, resp, err
	}
	return ca, resp, nil
}

// DeleteCustomUserAttribute removes the custom attribute of the specified user.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#delete-custom-attribute
func (s *CustomAttributesService) DeleteCustomUserAttribute(user int, key string, options ...RequestOptionFunc) (*Response, error) {
	return s.deleteCustomAttribute("users", user, key, options...)
}

// DeleteCustomGroupAttribute removes the custom attribute of the specified group.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#delete-custom-attribute
func (s *CustomAttributesService) DeleteCustomGroupAttribute(group int, key string, options ...RequestOptionFunc) (*Response, error) {
	return s.deleteCustomAttribute("groups", group, key, options...)
}

// DeleteCustomProjectAttribute removes the custom attribute of the specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/api/custom_attributes/#delete-custom-attribute
func (s *CustomAttributesService) DeleteCustomProjectAttribute(project int, key string, options ...RequestOptionFunc) (*Response, error) {
	return s.deleteCustomAttribute("projects", project, key, options...)
}

func (s *CustomAttributesService) deleteCustomAttribute(resource string, id int, key string, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("%s/%d/custom_attributes/%s", resource, id, key)
	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}
	return s.client.Do(req, nil)
}
