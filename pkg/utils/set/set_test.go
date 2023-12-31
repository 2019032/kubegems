// Copyright 2022 The kubegems.io Authors
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

package set

import (
	"testing"

	"kubegems.io/kubegems/pkg/utils/slice"
)

func TestSlice(t *testing.T) {
	tests := []struct {
		name       string
		elems      []string
		want       []string
		wantlength int
	}{
		{
			name:       "string",
			elems:      []string{"a", "a", "b", "c"},
			want:       []string{"a", "b", "c"},
			wantlength: 2,
		},
	}

	for _, tt := range tests {
		set := NewSet[string]()
		set.Append(tt.elems...)
		t.Run(tt.name, func(t *testing.T) {
			if got := set.Slice(); slice.SliceUniqueKey(got) != slice.SliceUniqueKey(tt.want) {
				t.Errorf("Slice() = %v, want %v", got, tt.want)
			}
			set.Remove("c")
			if set.Has("c") {
				t.Errorf("Has(\"c\") = true, want false")
			}
			if length := set.Len(); length != tt.wantlength {
				t.Errorf("Len() = %v, want %v", length, tt.wantlength)
			}
		})
	}
}
