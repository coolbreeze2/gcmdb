package runtime

import (
	"testing"
)

func TestNow(t *testing.T) {
	now := nowTime("")
	if len(now) != 14 {
		t.Errorf("Expected length of now to be 14, got %d", len(now))
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  map[string]any
		want     string
		wantErr  bool
	}{
		{
			name:     "simple variable",
			template: "Hello, ${name}!",
			context:  map[string]any{"name": "World"},
			want:     "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "missing variable",
			template: "Hello, ${name}!",
			context:  map[string]any{},
			want:     "Hello, !",
			wantErr:  false,
		},
		{
			name:     "empty template",
			template: "",
			context:  map[string]any{"foo": "bar"},
			want:     "",
			wantErr:  false,
		},
		{
			name:     "use now function no args error",
			template: "Now: ${now()}",
			context:  map[string]any{},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "use now function",
			template: "Now: ${now('')}",
			context:  map[string]any{},
			wantErr:  false,
		},
		{
			name:     "to_yaml filter",
			template: "Data: \n${ data | to_yaml }",
			context:  map[string]any{"data": map[string]any{"key": "value"}},
			want:     "Data: \nkey: value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate(tt.template, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.name != "use now function" && got != tt.want {
				t.Errorf("RenderTemplate() = %v, want %v", got, tt.want)
			}
			if tt.name == "use now function" && !tt.wantErr {
				if len(got) < 5 || got[:5] != "Now: " {
					t.Errorf("RenderTemplate() = %v, want prefix 'Now: '", got)
				}
			}
		})
	}
}
