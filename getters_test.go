package yamler

import (
	"testing"
)

func TestDocument_Get(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "get string",
			content: "key: value",
			path:    "key",
			want:    "value",
			wantErr: false,
		},
		{
			name:    "get int",
			content: "key: 123",
			path:    "key",
			want:    int64(123),
			wantErr: false,
		},
		{
			name:    "get float",
			content: "key: 123.45",
			path:    "key",
			want:    123.45,
			wantErr: false,
		},
		{
			name:    "get bool",
			content: "key: true",
			path:    "key",
			want:    true,
			wantErr: false,
		},
		{
			name:    "get array",
			content: "key: [1, 2, 3]",
			path:    "key",
			want:    []interface{}{int64(1), int64(2), int64(3)},
			wantErr: false,
		},
		{
			name:    "get map",
			content: "key:\n  nested: value",
			path:    "key",
			want:    map[string]interface{}{"nested": "value"},
			wantErr: false,
		},
		{
			name:    "get nested",
			content: "key:\n  nested: value",
			path:    "key.nested",
			want:    "value",
			wantErr: false,
		},
		{
			name:    "get non-existent",
			content: "key: value",
			path:    "nonexistent",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "get from non-map",
			content: "key: value",
			path:    "key.nested",
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

			got, err := doc.Get(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("Document.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_GetString(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "get string",
			content: "key: value",
			path:    "key",
			want:    "value",
			wantErr: false,
		},
		{
			name:    "get non-string",
			content: "key: 123",
			path:    "key",
			want:    "",
			wantErr: true,
		},
		{
			name:    "get non-existent",
			content: "key: value",
			path:    "nonexistent",
			want:    "",
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

			got, err := doc.GetString(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_GetInt(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    int64
		wantErr bool
	}{
		{
			name:    "get int",
			content: "key: 123",
			path:    "key",
			want:    123,
			wantErr: false,
		},
		{
			name:    "get string int",
			content: "key: \"123\"",
			path:    "key",
			want:    123,
			wantErr: false,
		},
		{
			name:    "get non-int",
			content: "key: value",
			path:    "key",
			want:    0,
			wantErr: true,
		},
		{
			name:    "get non-existent",
			content: "key: 123",
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

			got, err := doc.GetInt(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_GetFloat(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    float64
		wantErr bool
	}{
		{
			name:    "get float",
			content: "key: 123.45",
			path:    "key",
			want:    123.45,
			wantErr: false,
		},
		{
			name:    "get int as float",
			content: "key: 123",
			path:    "key",
			want:    123.0,
			wantErr: false,
		},
		{
			name:    "get string float",
			content: "key: \"123.45\"",
			path:    "key",
			want:    123.45,
			wantErr: false,
		},
		{
			name:    "get non-float",
			content: "key: value",
			path:    "key",
			want:    0,
			wantErr: true,
		},
		{
			name:    "get non-existent",
			content: "key: 123.45",
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

			got, err := doc.GetFloat(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.GetFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_GetBool(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    "get bool true",
			content: "key: true",
			path:    "key",
			want:    true,
			wantErr: false,
		},
		{
			name:    "get bool false",
			content: "key: false",
			path:    "key",
			want:    false,
			wantErr: false,
		},
		{
			name:    "get string true",
			content: "key: \"true\"",
			path:    "key",
			want:    true,
			wantErr: false,
		},
		{
			name:    "get string yes",
			content: "key: \"yes\"",
			path:    "key",
			want:    true,
			wantErr: false,
		},
		{
			name:    "get string 1",
			content: "key: \"1\"",
			path:    "key",
			want:    true,
			wantErr: false,
		},
		{
			name:    "get string on",
			content: "key: \"on\"",
			path:    "key",
			want:    true,
			wantErr: false,
		},
		{
			name:    "get string false",
			content: "key: \"false\"",
			path:    "key",
			want:    false,
			wantErr: false,
		},
		{
			name:    "get string no",
			content: "key: \"no\"",
			path:    "key",
			want:    false,
			wantErr: false,
		},
		{
			name:    "get string 0",
			content: "key: \"0\"",
			path:    "key",
			want:    false,
			wantErr: false,
		},
		{
			name:    "get string off",
			content: "key: \"off\"",
			path:    "key",
			want:    false,
			wantErr: false,
		},
		{
			name:    "get invalid string",
			content: "key: \"invalid\"",
			path:    "key",
			want:    false,
			wantErr: true,
		},
		{
			name:    "get non-bool",
			content: "key: 123",
			path:    "key",
			want:    false,
			wantErr: true,
		},
		{
			name:    "get non-existent",
			content: "key: true",
			path:    "nonexistent",
			want:    false,
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

			got, err := doc.GetBool(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Document.GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_GetSlice(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    []interface{}
		wantErr bool
	}{
		{
			name:    "get string slice",
			content: "key: [a, b, c]",
			path:    "key",
			want:    []interface{}{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "get int slice",
			content: "key: [1, 2, 3]",
			path:    "key",
			want:    []interface{}{int64(1), int64(2), int64(3)},
			wantErr: false,
		},
		{
			name:    "get mixed slice",
			content: "key: [1, true, string]",
			path:    "key",
			want:    []interface{}{int64(1), true, "string"},
			wantErr: false,
		},
		{
			name:    "get non-slice",
			content: "key: value",
			path:    "key",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "get non-existent",
			content: "key: [1, 2, 3]",
			path:    "nonexistent",
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

			got, err := doc.GetSlice(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("Document.GetSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_GetMap(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "get map",
			content: "key:\n  a: 1\n  b: true\n  c: string",
			path:    "key",
			want:    map[string]interface{}{"a": int64(1), "b": true, "c": "string"},
			wantErr: false,
		},
		{
			name:    "get nested map",
			content: "key:\n  nested:\n    a: 1\n    b: 2",
			path:    "key",
			want:    map[string]interface{}{"nested": map[string]interface{}{"a": int64(1), "b": int64(2)}},
			wantErr: false,
		},
		{
			name:    "get non-map",
			content: "key: value",
			path:    "key",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "get non-existent",
			content: "key:\n  a: 1",
			path:    "nonexistent",
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

			got, err := doc.GetMap(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.GetMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("Document.GetMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare values deeply
func deepEqual(a, b interface{}) bool {
	switch v := a.(type) {
	case []interface{}:
		w, ok := b.([]interface{})
		if !ok || len(v) != len(w) {
			return false
		}
		for i := range v {
			if !deepEqual(v[i], w[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		w, ok := b.(map[string]interface{})
		if !ok || len(v) != len(w) {
			return false
		}
		for k, val := range v {
			if !deepEqual(val, w[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
