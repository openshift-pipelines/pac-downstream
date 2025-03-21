//
// Copyright 2022, Ryan Glab <ryan.j.glab@gmail.com>
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
	"time"
)

// PersonalAccessTokensService handles communication with the personal access
// tokens related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/api/personal_access_tokens/
type PersonalAccessTokensService struct {
	client *Client
}

// PersonalAccessToken represents a personal access token.
//
// GitLab API docs: https://docs.gitlab.com/api/personal_access_tokens/
type PersonalAccessToken struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Revoked     bool       `json:"revoked"`
	CreatedAt   *time.Time `json:"created_at"`
	Description string     `json:"description"`
	Scopes      []string   `json:"scopes"`
	UserID      int        `json:"user_id"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	Active      bool       `json:"active"`
	ExpiresAt   *ISOTime   `json:"expires_at"`
	Token       string     `json:"token,omitempty"`
}

func (p PersonalAccessToken) String() string {
	return Stringify(p)
}

// ListPersonalAccessTokensOptions represents the available
// ListPersonalAccessTokens() options.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#list-all-personal-access-tokens
type ListPersonalAccessTokensOptions struct {
	ListOptions
	CreatedAfter   *ISOTime `url:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore  *ISOTime `url:"created_before,omitempty" json:"created_before,omitempty"`
	LastUsedAfter  *ISOTime `url:"last_used_after,omitempty" json:"last_used_after,omitempty"`
	LastUsedBefore *ISOTime `url:"last_used_before,omitempty" json:"last_used_before,omitempty"`
	Revoked        *bool    `url:"revoked,omitempty" json:"revoked,omitempty"`
	Search         *string  `url:"search,omitempty" json:"search,omitempty"`
	State          *string  `url:"state,omitempty" json:"state,omitempty"`
	UserID         *int     `url:"user_id,omitempty" json:"user_id,omitempty"`
}

// ListPersonalAccessTokens gets a list of all personal access tokens.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#list-all-personal-access-tokens
func (s *PersonalAccessTokensService) ListPersonalAccessTokens(opt *ListPersonalAccessTokensOptions, options ...RequestOptionFunc) ([]*PersonalAccessToken, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "personal_access_tokens", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pats []*PersonalAccessToken
	resp, err := s.client.Do(req, &pats)
	if err != nil {
		return nil, resp, err
	}

	return pats, resp, nil
}

// GetSinglePersonalAccessTokenByID get a single personal access token by its ID.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#get-details-on-a-personal-access-token
func (s *PersonalAccessTokensService) GetSinglePersonalAccessTokenByID(token int, options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	u := fmt.Sprintf("personal_access_tokens/%d", token)
	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pat := new(PersonalAccessToken)
	resp, err := s.client.Do(req, pat)
	if err != nil {
		return nil, resp, err
	}

	return pat, resp, nil
}

// GetSinglePersonalAccessToken get a single personal access token by using
// passing the token in a header.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#self-inform
func (s *PersonalAccessTokensService) GetSinglePersonalAccessToken(options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	u := "personal_access_tokens/self"
	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pat := new(PersonalAccessToken)
	resp, err := s.client.Do(req, pat)
	if err != nil {
		return nil, resp, err
	}

	return pat, resp, nil
}

// RotatePersonalAccessTokenOptions represents the available RotatePersonalAccessToken()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#rotate-a-personal-access-token
type RotatePersonalAccessTokenOptions struct {
	ExpiresAt *ISOTime `url:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// RotatePersonalAccessToken is a backwards-compat shim for RotatePersonalAccessTokenByID.
func (s *PersonalAccessTokensService) RotatePersonalAccessToken(token int, opt *RotatePersonalAccessTokenOptions, options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	return s.RotatePersonalAccessTokenByID(token, opt, options...)
}

// RotatePersonalAccessTokenByID revokes a token and returns a new token that
// expires in one week per default.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#rotate-a-personal-access-token
func (s *PersonalAccessTokensService) RotatePersonalAccessTokenByID(token int, opt *RotatePersonalAccessTokenOptions, options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	u := fmt.Sprintf("personal_access_tokens/%d/rotate", token)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pat := new(PersonalAccessToken)
	resp, err := s.client.Do(req, pat)
	if err != nil {
		return nil, resp, err
	}

	return pat, resp, nil
}

// RotatePersonalAccessTokenSelf revokes the currently authenticated token
// and returns a new token that expires in one week per default.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#self-rotate
func (s *PersonalAccessTokensService) RotatePersonalAccessTokenSelf(opt *RotatePersonalAccessTokenOptions, options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	u := "personal_access_tokens/self/rotate"

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pat := new(PersonalAccessToken)
	resp, err := s.client.Do(req, pat)
	if err != nil {
		return nil, resp, err
	}

	return pat, resp, nil
}

// RevokePersonalAccessToken is a backwards-compat shim for RevokePersonalAccessTokenByID.
func (s *PersonalAccessTokensService) RevokePersonalAccessToken(token int, options ...RequestOptionFunc) (*Response, error) {
	return s.RevokePersonalAccessTokenByID(token, options...)
}

// RevokePersonalAccessTokenByID revokes a personal access token by its ID.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#revoke-a-personal-access-token
func (s *PersonalAccessTokensService) RevokePersonalAccessTokenByID(token int, options ...RequestOptionFunc) (*Response, error) {
	u := fmt.Sprintf("personal_access_tokens/%d", token)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// RevokePersonalAccessTokenSelf revokes the currently authenticated
// personal access token.
//
// GitLab API docs:
// https://docs.gitlab.com/api/personal_access_tokens/#self-revoke
func (s *PersonalAccessTokensService) RevokePersonalAccessTokenSelf(options ...RequestOptionFunc) (*Response, error) {
	u := "personal_access_tokens/self"

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
