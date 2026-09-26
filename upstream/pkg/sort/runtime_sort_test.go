/*
Copyright 2014 The Kubernetes Authors.

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

package sort

import (
	"reflect"
	gsort "sort"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

func createPodSpecResource(t *testing.T, memReq, memLimit, cpuReq, cpuLimit string) corev1.PodSpec {
	t.Helper()
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{},
					Limits:   corev1.ResourceList{},
				},
			},
		},
	}

	req := podSpec.Containers[0].Resources.Requests
	if memReq != "" {
		memReq, err := resource.ParseQuantity(memReq)
		assert.NilError(t, err, "memory request string is not a valid quantity")
		req["memory"] = memReq
	}
	if cpuReq != "" {
		cpuReq, err := resource.ParseQuantity(cpuReq)
		assert.NilError(t, err, "cpu request string is not a valid quantity")
		req["cpu"] = cpuReq
	}
	limit := podSpec.Containers[0].Resources.Limits
	if memLimit != "" {
		memLimit, err := resource.ParseQuantity(memLimit)
		assert.NilError(t, err, "memory limit string is not a valid quantity")
		limit["memory"] = memLimit
	}
	if cpuLimit != "" {
		cpuLimit, err := resource.ParseQuantity(cpuLimit)
		assert.NilError(t, err, "cpu limit string is not a valid quantity")
		limit["cpu"] = cpuLimit
	}

	return podSpec
}

func TestRuntimeSortLess(t *testing.T) {
	testobj := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "b",
				},
				Spec: createPodSpecResource(t, "0.5", "", "1Gi", ""),
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "c",
				},
				Spec: createPodSpecResource(t, "2", "", "1Ti", ""),
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "a",
				},
				Spec: createPodSpecResource(t, "10m", "", "1Ki", ""),
			},
		},
	}

	testobjs, err := meta.ExtractList(testobj)
	assert.NilError(t, err)

	testfieldName := "{.metadata.name}"
	testruntimeSortName := NewRuntimeSort(testfieldName, testobjs)

	testfieldCPU := "{.spec.containers[].resources.requests.cpu}"
	testruntimeSortCPU := NewRuntimeSort(testfieldCPU, testobjs)

	testfieldMemory := "{.spec.containers[].resources.requests.memory}"
	testruntimeSortMemory := NewRuntimeSort(testfieldMemory, testobjs)

	tests := []struct {
		name         string
		runtimeSort  *RuntimeSort
		i            int
		j            int
		expectResult bool
		expectErr    bool
	}{
		{
			name:         "test name b c less true",
			runtimeSort:  testruntimeSortName,
			i:            0,
			j:            1,
			expectResult: true,
		},
		{
			name:         "test name c a less false",
			runtimeSort:  testruntimeSortName,
			i:            1,
			j:            2,
			expectResult: false,
		},
		{
			name:         "test name b a less false",
			runtimeSort:  testruntimeSortName,
			i:            0,
			j:            2,
			expectResult: false,
		},
		{
			name:         "test cpu 0.5 2 less true",
			runtimeSort:  testruntimeSortCPU,
			i:            0,
			j:            1,
			expectResult: true,
		},
		{
			name:         "test cpu 2 10mi less false",
			runtimeSort:  testruntimeSortCPU,
			i:            1,
			j:            2,
			expectResult: false,
		},
		{
			name:         "test cpu 0.5 10mi less false",
			runtimeSort:  testruntimeSortCPU,
			i:            0,
			j:            2,
			expectResult: false,
		},
		{
			name:         "test memory 1Gi 1Ti less true",
			runtimeSort:  testruntimeSortMemory,
			i:            0,
			j:            1,
			expectResult: true,
		},
		{
			name:         "test memory 1Ti 1Ki less false",
			runtimeSort:  testruntimeSortMemory,
			i:            1,
			j:            2,
			expectResult: false,
		},
		{
			name:         "test memory 1Gi 1Ki less false",
			runtimeSort:  testruntimeSortMemory,
			i:            0,
			j:            2,
			expectResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.runtimeSort.Less(test.i, test.j)
			assert.Equal(t, test.expectResult, result)
		})
	}
}

func TestIsLess(t *testing.T) {
	interfaceValues := func(left, right any) (reflect.Value, reflect.Value) {
		values := []any{left, right}
		return reflect.ValueOf(values).Index(0), reflect.ValueOf(values).Index(1)
	}

	type fallbackStruct struct {
		Number int
	}

	tests := []struct {
		name    string
		values  func(t *testing.T) (reflect.Value, reflect.Value)
		want    bool
		wantErr string
	}{
		{
			name: "int kind",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(-1), reflect.ValueOf(1)
			},
			want: true,
		},
		{
			name: "uint kind",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(uint(1)), reflect.ValueOf(uint(2))
			},
			want: true,
		},
		{
			name: "float kind",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(1.5), reflect.ValueOf(2.5)
			},
			want: true,
		},
		{
			name: "string natural sort",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf("item2"), reflect.ValueOf("item10")
			},
			want: true,
		},
		{
			name: "pointer",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				left := 1
				right := 2
				return reflect.ValueOf(&left), reflect.ValueOf(&right)
			},
			want: true,
		},
		{
			name: "metav1 time struct",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(metav1.NewTime(time.Unix(1, 0))), reflect.ValueOf(metav1.NewTime(time.Unix(2, 0)))
			},
			want: true,
		},
		{
			name: "resource quantity struct",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(resource.MustParse("1Gi")), reflect.ValueOf(resource.MustParse("2Gi"))
			},
			want: true,
		},
		{
			name: "generic empty struct equal returns true",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(struct{}{}), reflect.ValueOf(struct{}{})
			},
			want: true,
		},
		{
			name: "generic struct first field less",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(fallbackStruct{Number: 1}), reflect.ValueOf(fallbackStruct{Number: 2})
			},
			want: true,
		},
		{
			name: "generic struct error inside field",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(struct{ Unsupported bool }{Unsupported: false}), reflect.ValueOf(struct{ Unsupported bool }{Unsupported: true})
			},
			wantErr: "unsortable type",
		},
		{
			name: "array different lengths",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf([0]int{}), reflect.ValueOf([1]int{1})
			},
			want: true,
		},
		{
			name: "slice elem not less",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf([]int{2}), reflect.ValueOf([]int{1})
			},
			want: false,
		},
		{
			name: "interface both nil",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(nil, nil)
			},
			want: false,
		},
		{
			name: "interface nil non nil",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(nil, 1)
			},
			want: true,
		},
		{
			name: "interface value nil",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(1, nil)
			},
			want: false,
		},
		{
			name: "interface uint8",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(uint8(1), uint8(2))
			},
			want: true,
		},
		{
			name: "interface uint16",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(uint16(1), uint16(2))
			},
			want: true,
		},
		{
			name: "interface uint32",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(uint32(1), uint32(2))
			},
			want: true,
		},
		{
			name: "interface uint64",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(uint64(1), uint64(2))
			},
			want: true,
		},
		{
			name: "interface int8",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(int8(1), int8(2))
			},
			want: true,
		},
		{
			name: "interface int16",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(int16(1), int16(2))
			},
			want: true,
		},
		{
			name: "interface int32",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(int32(1), int32(2))
			},
			want: true,
		},
		{
			name: "interface int64",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(int64(1), int64(2))
			},
			want: true,
		},
		{
			name: "interface uint",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(uint(1), uint(2))
			},
			want: true,
		},
		{
			name: "interface int",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(1, 2)
			},
			want: true,
		},
		{
			name: "interface float32",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(float32(1.5), float32(2.5))
			},
			want: true,
		},
		{
			name: "interface float64",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(1.5, 2.5)
			},
			want: true,
		},
		{
			name: "interface string quantities",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues("1Gi", "2Gi")
			},
			want: true,
		},
		{
			name: "interface string one not quantity",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues("item2", "item10")
			},
			want: true,
		},
		{
			name: "interface mismatched types",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(uint8(1), uint16(2))
			},
			wantErr: "unsortable interface",
		},
		{
			name: "interface unsupported type",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return interfaceValues(struct{}{}, struct{}{})
			},
			wantErr: "unsortable type",
		},
		{
			name: "default unsupported bool",
			values: func(t *testing.T) (reflect.Value, reflect.Value) {
				t.Helper()
				return reflect.ValueOf(true), reflect.ValueOf(false)
			},
			wantErr: "unsortable type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, right := tt.values(t)
			got, err := isLess(left, right)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOriginalPosition(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		objects  []runtime.Object
		expected []int
	}{
		{
			name:  "tracks original positions after sorting",
			field: "{.metadata.name}",
			objects: []runtime.Object{
				&unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "c"}}},
				&unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "a"}}},
				&unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "b"}}},
			},
			expected: []int{1, 2, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewRuntimeSort(tt.field, tt.objects)
			gsort.Sort(sorter)

			for idx, want := range tt.expected {
				assert.Equal(t, want, sorter.OriginalPosition(idx))
			}
			assert.Equal(t, -1, sorter.OriginalPosition(-1))
			assert.Equal(t, -1, sorter.OriginalPosition(len(tt.objects)))
		})
	}
}

func TestFindJSONPathResultsError(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		object  runtime.Object
		wantErr string
	}{
		{
			name:    "array index against scalar",
			field:   "{.metadata.name[1]}",
			object:  &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "pod"}}},
			wantErr: "not array or slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := jsonpath.New("sorting").AllowMissingKeys(true)
			err := parser.Parse(tt.field)
			assert.NilError(t, err)

			_, err = findJSONPathResults(parser, tt.object)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
