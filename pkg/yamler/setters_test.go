package yamler

import (
	"testing"
)

func TestDocument_Set(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "set string",
			content: "key: value",
			path:    "key",
			value:   "new value",
			want:    "key: new value\n",
			wantErr: false,
		},
		{
			name:    "set int",
			content: "key: value",
			path:    "key",
			value:   123,
			want:    "key: 123\n",
			wantErr: false,
		},
		{
			name:    "set float",
			content: "key: value",
			path:    "key",
			value:   123.45,
			want:    "key: 123.45\n",
			wantErr: false,
		},
		{
			name:    "set bool",
			content: "key: value",
			path:    "key",
			value:   true,
			want:    "key: true\n",
			wantErr: false,
		},
		{
			name:    "set slice",
			content: "key: value",
			path:    "key",
			value:   []interface{}{1, "two", true},
			want:    "key:\n  - 1\n  - two\n  - true\n",
			wantErr: false,
		},
		{
			name:    "set map",
			content: "key: value",
			path:    "key",
			value:   map[string]interface{}{"a": 1, "b": "two"},
			want:    "key:\n  a: 1\n  b: two\n",
			wantErr: false,
		},
		{
			name:    "set nested",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   "new value",
			want:    "key:\n  nested: new value\n",
			wantErr: false,
		},
		{
			name:    "set new nested",
			content: "key: value",
			path:    "key.nested",
			value:   "new value",
			want:    "key:\n  nested: new value\n",
			wantErr: false,
		},
		{
			name:    "set root",
			content: "key: value",
			path:    "",
			value:   map[string]interface{}{"new": "value"},
			want:    "new: value\n",
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

			err = doc.Set(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.Set() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetString(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   string
		want    string
		wantErr bool
	}{
		{
			name:    "set string",
			content: "key: value",
			path:    "key",
			value:   "new value",
			want:    "key: new value\n",
			wantErr: false,
		},
		{
			name:    "set nested string",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   "new value",
			want:    "key:\n  nested: new value\n",
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

			err = doc.SetString(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetString() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetInt(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   int64
		want    string
		wantErr bool
	}{
		{
			name:    "set int",
			content: "key: value",
			path:    "key",
			value:   123,
			want:    "key: 123\n",
			wantErr: false,
		},
		{
			name:    "set nested int",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   123,
			want:    "key:\n  nested: 123\n",
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

			err = doc.SetInt(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetInt() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetFloat(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   float64
		want    string
		wantErr bool
	}{
		{
			name:    "set float",
			content: "key: value",
			path:    "key",
			value:   123.45,
			want:    "key: 123.45\n",
			wantErr: false,
		},
		{
			name:    "set nested float",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   123.45,
			want:    "key:\n  nested: 123.45\n",
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

			err = doc.SetFloat(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetFloat() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetBool(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   bool
		want    string
		wantErr bool
	}{
		{
			name:    "set bool true",
			content: "key: value",
			path:    "key",
			value:   true,
			want:    "key: true\n",
			wantErr: false,
		},
		{
			name:    "set bool false",
			content: "key: value",
			path:    "key",
			value:   false,
			want:    "key: false\n",
			wantErr: false,
		},
		{
			name:    "set nested bool",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   true,
			want:    "key:\n  nested: true\n",
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

			err = doc.SetBool(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetBool() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   []string
		want    string
		wantErr bool
	}{
		{
			name:    "set string slice",
			content: "key: value",
			path:    "key",
			value:   []string{"a", "b", "c"},
			want:    "key:\n  - a\n  - b\n  - c\n",
			wantErr: false,
		},
		{
			name:    "set nested string slice",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   []string{"a", "b", "c"},
			want:    "key:\n  nested:\n    - a\n    - b\n    - c\n",
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

			err = doc.SetStringSlice(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetStringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetStringSlice() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetIntSlice(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   []int64
		want    string
		wantErr bool
	}{
		{
			name:    "set int slice",
			content: "key: value",
			path:    "key",
			value:   []int64{1, 2, 3},
			want:    "key:\n  - 1\n  - 2\n  - 3\n",
			wantErr: false,
		},
		{
			name:    "set nested int slice",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   []int64{1, 2, 3},
			want:    "key:\n  nested:\n    - 1\n    - 2\n    - 3\n",
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

			err = doc.SetIntSlice(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetIntSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetIntSlice() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetFloatSlice(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   []float64
		want    string
		wantErr bool
	}{
		{
			name:    "set float slice",
			content: "key: value",
			path:    "key",
			value:   []float64{1.1, 2.2, 3.3},
			want:    "key:\n  - 1.1\n  - 2.2\n  - 3.3\n",
			wantErr: false,
		},
		{
			name:    "set nested float slice",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   []float64{1.1, 2.2, 3.3},
			want:    "key:\n  nested:\n    - 1.1\n    - 2.2\n    - 3.3\n",
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

			err = doc.SetFloatSlice(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetFloatSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetFloatSlice() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetBoolSlice(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   []bool
		want    string
		wantErr bool
	}{
		{
			name:    "set bool slice",
			content: "key: value",
			path:    "key",
			value:   []bool{true, false, true},
			want:    "key:\n  - true\n  - false\n  - true\n",
			wantErr: false,
		},
		{
			name:    "set nested bool slice",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value:   []bool{true, false, true},
			want:    "key:\n  nested:\n    - true\n    - false\n    - true\n",
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

			err = doc.SetBoolSlice(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetBoolSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetBoolSlice() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDocument_SetMapSlice(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		value   []map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "set map slice",
			content: "key: value",
			path:    "key",
			value: []map[string]interface{}{
				{"a": 1, "b": "two"},
				{"c": true, "d": 4.5},
			},
			want:    "key:\n  - a: 1\n    b: two\n  - c: true\n    d: 4.5\n",
			wantErr: false,
		},
		{
			name:    "set nested map slice",
			content: "key:\n  nested: value",
			path:    "key.nested",
			value: []map[string]interface{}{
				{"a": 1, "b": "two"},
				{"c": true, "d": 4.5},
			},
			want:    "key:\n  nested:\n    - a: 1\n      b: two\n    - c: true\n      d: 4.5\n",
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

			err = doc.SetMapSlice(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.SetMapSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := doc.String()
				if err != nil {
					t.Errorf("Document.String() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Document.SetMapSlice() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
