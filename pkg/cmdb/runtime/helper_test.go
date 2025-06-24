package runtime

import (
	"gcmdb/pkg/cmdb"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFieldValueByTagEmpty(t *testing.T) {
	app := *cmdb.NewApp()
	v := reflect.ValueOf(app)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(t, []TagValuePair{}, result)
}

func TestGetFieldValueByTag(t *testing.T) {
	app := &cmdb.Datacenter{Spec: cmdb.DatacenterSpec{PrivateKey: "test"}}
	v := reflect.ValueOf(app)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(t, []TagValuePair{{"Secret", "test", "spec.privateKey"}}, result)
}

func TestGetFieldValueByTagMap(t *testing.T) {
	m := map[string]string{}
	v := reflect.ValueOf(m)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(t, []TagValuePair{}, result)
}

func TestGetFieldValueByTagList(t *testing.T) {
	type Case struct {
		Field1 []string `reference:"Secret"`
		Field2 []string `reference:"Datacenter"`
	}
	c := Case{Field1: []string{"v1", "v1", "v2"}, Field2: []string{"v1", "v1", "v2"}}
	v := reflect.ValueOf(c)
	result := GetFieldValueByTag(v, "", "reference")
	assert.Equal(
		t,
		[]TagValuePair{
			{"Secret", "v1", "Field1"},
			{"Secret", "v2", "Field1"},
			{"Datacenter", "v1", "Field2"},
			{"Datacenter", "v2", "Field2"},
		},
		result,
	)
}

func TestRecSetItem_SingleLevel(t *testing.T) {
	obj := make(map[string]any)
	RecSetItem(obj, "foo", 123)
	want := map[string]any{"foo": 123}
	if !reflect.DeepEqual(obj, want) {
		t.Errorf("got %v, want %v", obj, want)
	}
}

func TestRecSetItem_MultiLevel_NewPath(t *testing.T) {
	obj := make(map[string]any)
	RecSetItem(obj, "foo.bar", "baz")
	want := map[string]any{
		"foo": map[string]any{
			"bar": "baz",
		},
	}
	if !reflect.DeepEqual(obj, want) {
		t.Errorf("got %v, want %v", obj, want)
	}
}

func TestRecSetItem_MultiLevel_ExistingMap(t *testing.T) {
	obj := map[string]any{
		"foo": map[string]any{
			"old": 1,
		},
	}
	RecSetItem(obj, "foo.bar", 2)
	want := map[string]any{
		"foo": map[string]any{
			"old": 1,
			"bar": 2,
		},
	}
	if !reflect.DeepEqual(obj, want) {
		t.Errorf("got %v, want %v", obj, want)
	}
}

func TestRecSetItem_OverwriteNonMap(t *testing.T) {
	obj := map[string]any{
		"foo": 123,
	}
	RecSetItem(obj, "foo.bar", "baz")
	want := map[string]any{
		"foo": map[string]any{
			"bar": "baz",
		},
	}
	if !reflect.DeepEqual(obj, want) {
		t.Errorf("got %v, want %v", obj, want)
	}
}

func TestRecSetItem_DeeplyNested(t *testing.T) {
	obj := make(map[string]any)
	RecSetItem(obj, "a.b.c.d", 42)
	want := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": 42,
				},
			},
		},
	}
	if !reflect.DeepEqual(obj, want) {
		t.Errorf("got %v, want %v", obj, want)
	}
}

func TestRecSetItem_EmptyPath(t *testing.T) {
	obj := make(map[string]any)
	RecSetItem(obj, "", "val")
	want := map[string]any{
		"": "val",
	}
	if !reflect.DeepEqual(obj, want) {
		t.Errorf("got %v, want %v", obj, want)
	}
}

func TestMerge2Dict_BothEmpty(t *testing.T) {
	current := map[string]any{}
	target := map[string]any{}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestMerge2Dict_TargetAddsKeys(t *testing.T) {
	current := map[string]any{"a": 1}
	target := map[string]any{"b": 2}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{"a": 1, "b": 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestMerge2Dict_OverwriteNilWithAllowedType(t *testing.T) {
	current := map[string]any{"a": nil}
	target := map[string]any{"a": 123}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{"a": 123}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestMerge2Dict_ConflictNotOverwritten(t *testing.T) {
	current := map[string]any{"a": 1}
	target := map[string]any{"a": 2}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{"a": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestMerge2Dict_RecursiveMerge(t *testing.T) {
	current := map[string]any{
		"outer": map[string]any{
			"x": 1,
		},
	}
	target := map[string]any{
		"outer": map[string]any{
			"y": 2,
		},
	}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{
		"outer": map[string]any{
			"x": 1,
			"y": 2,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestMerge2Dict_RecursiveConflict(t *testing.T) {
	current := map[string]any{
		"outer": map[string]any{
			"x": 1,
		},
	}
	target := map[string]any{
		"outer": map[string]any{
			"x": 2,
		},
	}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{
		"outer": map[string]any{
			"x": 1,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestMerge2Dict_NilCurrentMap(t *testing.T) {
	var current map[string]any
	target := map[string]any{"a": 1}
	// Should not panic, but current is nil so cannot merge in place
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when current is nil")
		}
	}()
	Merge2Dict(current, target, nil)
}

func TestMerge2Dict_SliceAllowedType(t *testing.T) {
	current := map[string]any{"a": nil}
	target := map[string]any{"a": []any{1, 2}}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{"a": []any{1, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestMerge2Dict_DeepNested(t *testing.T) {
	current := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"x": 1,
			},
		},
	}
	target := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"y": 2,
			},
		},
	}
	got := Merge2Dict(current, target, nil)
	want := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"x": 1,
				"y": 2,
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestIsAllowedType(t *testing.T) {
	assert.Equal(t, isAllowedType(map[string]string{}), false)
}
