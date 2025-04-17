/*
Copyright 2020 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"

	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	scheme "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// StepActionsGetter has a method to return a StepActionInterface.
// A group's client should implement this interface.
type StepActionsGetter interface {
	StepActions(namespace string) StepActionInterface
}

// StepActionInterface has methods to work with StepAction resources.
type StepActionInterface interface {
	Create(ctx context.Context, stepAction *v1alpha1.StepAction, opts v1.CreateOptions) (*v1alpha1.StepAction, error)
	Update(ctx context.Context, stepAction *v1alpha1.StepAction, opts v1.UpdateOptions) (*v1alpha1.StepAction, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.StepAction, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.StepActionList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.StepAction, err error)
	StepActionExpansion
}

// stepActions implements StepActionInterface
type stepActions struct {
	*gentype.ClientWithList[*v1alpha1.StepAction, *v1alpha1.StepActionList]
}

// newStepActions returns a StepActions
func newStepActions(c *TektonV1alpha1Client, namespace string) *stepActions {
	return &stepActions{
		gentype.NewClientWithList[*v1alpha1.StepAction, *v1alpha1.StepActionList](
			"stepactions",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.StepAction { return &v1alpha1.StepAction{} },
			func() *v1alpha1.StepActionList { return &v1alpha1.StepActionList{} }),
	}
}
