package yamler

import (
	"testing"
)

func TestDocument_GetArrayLength(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    int
		wantErr bool
	}{
		{
			name:    "get empty array length",
			content: "key: []",
			path:    "key",
			want:    0,
			wantErr: false,
		},
		{
			name:    "get array length",
			content: "key: [1, 2, 3]",
			path:    "key",
			want:    3,
			wantErr: false,
		},
		{
			name:    "get nested array length",
			content: "key:\n  nested: [1, 2, 3, 4]",
			path:    "key.nested",
			want:    4,
			wantErr: false,
		},
		{
			name:    "get non-array length",
			content: "key: value",
			path:    "key",
			want:    0,
			wantErr: true,
		},
		{
			name:    "get non-existent array length",
			content: "key: [1, 2, 3]",
			path:    "nonexistent",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			got, err := doc.GetArrayLength(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetArrayLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.GetArrayLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_AppendToArray(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "append to empty array",
			content: "key: []",
			path:    "key",
			value:   "value",
			want:    "key: [value]\n",
			wantErr: false,
		},
		{
			name:    "append to array",
			content: "key: [1, 2]",
			path:    "key",
			value:   3,
			want:    "key: [1, 2, 3]\n",
			wantErr: false,
		},
		{
			name:    "append to nested array",
			content: "key:\n  nested: [1, 2]",
			path:    "key.nested",
			value:   3,
			want:    "key:\n  nested: [1, 2, 3]\n",
			wantErr: false,
		},
		{
			name:    "append to non-array",
			content: "key: value",
			path:    "key",
			value:   "new",
			want:    "key: value\n",
			wantErr: true,
		},
		{
			name:    "append to non-existent path",
			content: "key: [1, 2, 3]",
			path:    "nonexistent",
			value:   4,
			want:    "key: [1, 2, 3]\nnonexistent: [4]\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			err = doc.AppendToArray(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.AppendToArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.AppendToArray() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_RemoveFromArray(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		index   int
		want    string
		wantErr bool
	}{
		{
			name:    "remove first element",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   0,
			want:    "key: [2, 3]\n",
			wantErr: false,
		},
		{
			name:    "remove middle element",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   1,
			want:    "key: [1, 3]\n",
			wantErr: false,
		},
		{
			name:    "remove last element",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   2,
			want:    "key: [1, 2]\n",
			wantErr: false,
		},
		{
			name:    "remove from nested array",
			content: "key:\n  nested: [1, 2, 3]",
			path:    "key.nested",
			index:   1,
			want:    "key:\n  nested: [1, 3]\n",
			wantErr: false,
		},
		{
			name:    "remove from non-array",
			content: "key: value",
			path:    "key",
			index:   0,
			want:    "key: value\n",
			wantErr: true,
		},
		{
			name:    "remove from non-existent path",
			content: "key: [1, 2, 3]",
			path:    "nonexistent",
			index:   0,
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
		{
			name:    "remove with negative index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   -1,
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
		{
			name:    "remove with out of bounds index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   3,
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			err = doc.RemoveFromArray(tt.path, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.RemoveFromArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.RemoveFromArray() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_UpdateArrayElement(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		index   int
		value   interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "update first element",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   0,
			value:   "new",
			want:    "key: [new, 2, 3]\n",
			wantErr: false,
		},
		{
			name:    "update middle element",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   1,
			value:   "new",
			want:    "key: [1, new, 3]\n",
			wantErr: false,
		},
		{
			name:    "update last element",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   2,
			value:   "new",
			want:    "key: [1, 2, new]\n",
			wantErr: false,
		},
		{
			name:    "update nested array element",
			content: "key:\n  nested: [1, 2, 3]",
			path:    "key.nested",
			index:   1,
			value:   "new",
			want:    "key:\n  nested: [1, new, 3]\n",
			wantErr: false,
		},
		{
			name:    "update non-array",
			content: "key: value",
			path:    "key",
			index:   0,
			value:   "new",
			want:    "key: value\n",
			wantErr: true,
		},
		{
			name:    "update non-existent path",
			content: "key: [1, 2, 3]",
			path:    "nonexistent",
			index:   0,
			value:   "new",
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
		{
			name:    "update with negative index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   -1,
			value:   "new",
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
		{
			name:    "update with out of bounds index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   3,
			value:   "new",
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			err = doc.UpdateArrayElement(tt.path, tt.index, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.UpdateArrayElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.UpdateArrayElement() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_InsertIntoArray(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		index   int
		value   interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "insert at beginning",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   0,
			value:   "new",
			want:    "key: [new, 1, 2, 3]\n",
			wantErr: false,
		},
		{
			name:    "insert in middle",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   1,
			value:   "new",
			want:    "key: [1, new, 2, 3]\n",
			wantErr: false,
		},
		{
			name:    "insert at end",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   3,
			value:   "new",
			want:    "key: [1, 2, 3, new]\n",
			wantErr: false,
		},
		{
			name:    "insert into nested array",
			content: "key:\n  nested: [1, 2, 3]",
			path:    "key.nested",
			index:   1,
			value:   "new",
			want:    "key:\n  nested: [1, new, 2, 3]\n",
			wantErr: false,
		},
		{
			name:    "insert into non-array",
			content: "key: value",
			path:    "key",
			index:   0,
			value:   "new",
			want:    "key: value\n",
			wantErr: true,
		},
		{
			name:    "insert into non-existent path",
			content: "key: [1, 2, 3]",
			path:    "nonexistent",
			index:   0,
			value:   "new",
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
		{
			name:    "insert with negative index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   -1,
			value:   "new",
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
		{
			name:    "insert with out of bounds index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   4,
			value:   "new",
			want:    "key: [1, 2, 3]\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			err = doc.InsertIntoArray(tt.path, tt.index, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.InsertIntoArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.InsertIntoArray() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_GetArrayElement(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		index   int
		want    interface{}
		wantErr bool
	}{
		{
			name:    "get string element",
			content: "key: [a, b, c]",
			path:    "key",
			index:   1,
			want:    "b",
			wantErr: false,
		},
		{
			name:    "get int element",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   1,
			want:    int64(2),
			wantErr: false,
		},
		{
			name:    "get float element",
			content: "key: [1.1, 2.2, 3.3]",
			path:    "key",
			index:   1,
			want:    2.2,
			wantErr: false,
		},
		{
			name:    "get bool element",
			content: "key: [true, false, true]",
			path:    "key",
			index:   1,
			want:    false,
			wantErr: false,
		},
		{
			name:    "get from nested array",
			content: "key:\n  nested: [1, 2, 3]",
			path:    "key.nested",
			index:   1,
			want:    int64(2),
			wantErr: false,
		},
		{
			name:    "get from non-array",
			content: "key: value",
			path:    "key",
			index:   0,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "get from non-existent path",
			content: "key: [1, 2, 3]",
			path:    "nonexistent",
			index:   0,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "get with negative index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   -1,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "get with out of bounds index",
			content: "key: [1, 2, 3]",
			path:    "key",
			index:   3,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			got, err := doc.GetArrayElement(tt.path, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetArrayElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("Document.GetArrayElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_GetTypedArrayElement(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		path       string
		index      int
		targetType string
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "get string element",
			content:    "key: [a, b, c]",
			path:       "key",
			index:      1,
			targetType: "string",
			want:       "b",
			wantErr:    false,
		},
		{
			name:       "get int element",
			content:    "key: [1, 2, 3]",
			path:       "key",
			index:      1,
			targetType: "int",
			want:       int64(2),
			wantErr:    false,
		},
		{
			name:       "get float element",
			content:    "key: [1.1, 2.2, 3.3]",
			path:       "key",
			index:      1,
			targetType: "float",
			want:       2.2,
			wantErr:    false,
		},
		{
			name:       "get bool element",
			content:    "key: [true, false, true]",
			path:       "key",
			index:      1,
			targetType: "bool",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "get string as int",
			content:    "key: [\"1\", \"2\", \"3\"]",
			path:       "key",
			index:      1,
			targetType: "int",
			want:       int64(2),
			wantErr:    false,
		},
		{
			name:       "get string as float",
			content:    "key: [\"1.1\", \"2.2\", \"3.3\"]",
			path:       "key",
			index:      1,
			targetType: "float",
			want:       2.2,
			wantErr:    false,
		},
		{
			name:       "get string as bool",
			content:    "key: [\"true\", \"false\", \"true\"]",
			path:       "key",
			index:      1,
			targetType: "bool",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "get invalid type",
			content:    "key: [1, 2, 3]",
			path:       "key",
			index:      1,
			targetType: "invalid",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "get from non-array",
			content:    "key: value",
			path:       "key",
			index:      0,
			targetType: "string",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "get from non-existent path",
			content:    "key: [1, 2, 3]",
			path:       "nonexistent",
			index:      0,
			targetType: "int",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "get with negative index",
			content:    "key: [1, 2, 3]",
			path:       "key",
			index:      -1,
			targetType: "int",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "get with out of bounds index",
			content:    "key: [1, 2, 3]",
			path:       "key",
			index:      3,
			targetType: "int",
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			got, err := doc.GetTypedArrayElement(tt.path, tt.index, tt.targetType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetTypedArrayElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("Document.GetTypedArrayElement() = %v, want %v", got, tt.want)
			}
		})
	}
}
