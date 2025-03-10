/*
Copyright The Kubernetes Authors.

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

package fake

import (
	v1alpha1 "k8s.io/api/storagemigration/v1alpha1"
	storagemigrationv1alpha1 "k8s.io/client-go/applyconfigurations/storagemigration/v1alpha1"
	gentype "k8s.io/client-go/gentype"
	typedstoragemigrationv1alpha1 "k8s.io/client-go/kubernetes/typed/storagemigration/v1alpha1"
)

// fakeStorageVersionMigrations implements StorageVersionMigrationInterface
type fakeStorageVersionMigrations struct {
	*gentype.FakeClientWithListAndApply[*v1alpha1.StorageVersionMigration, *v1alpha1.StorageVersionMigrationList, *storagemigrationv1alpha1.StorageVersionMigrationApplyConfiguration]
	Fake *FakeStoragemigrationV1alpha1
}

func newFakeStorageVersionMigrations(fake *FakeStoragemigrationV1alpha1) typedstoragemigrationv1alpha1.StorageVersionMigrationInterface {
	return &fakeStorageVersionMigrations{
		gentype.NewFakeClientWithListAndApply[*v1alpha1.StorageVersionMigration, *v1alpha1.StorageVersionMigrationList, *storagemigrationv1alpha1.StorageVersionMigrationApplyConfiguration](
			fake.Fake,
			"",
			v1alpha1.SchemeGroupVersion.WithResource("storageversionmigrations"),
			v1alpha1.SchemeGroupVersion.WithKind("StorageVersionMigration"),
			func() *v1alpha1.StorageVersionMigration { return &v1alpha1.StorageVersionMigration{} },
			func() *v1alpha1.StorageVersionMigrationList { return &v1alpha1.StorageVersionMigrationList{} },
			func(dst, src *v1alpha1.StorageVersionMigrationList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.StorageVersionMigrationList) []*v1alpha1.StorageVersionMigration {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.StorageVersionMigrationList, items []*v1alpha1.StorageVersionMigration) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
