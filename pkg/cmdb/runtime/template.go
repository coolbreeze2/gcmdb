package runtime

import (
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

func RenderTemplate(template string, context map[string]any) (string, error) {
	// Create a new template from the string
	tpl, err := gonja.FromString(template)
	if err != nil {
		return "", err
	}

	// Create a new execution context with the provided context
	ctx := exec.NewContext(context)

	// Render the template with the context
	return tpl.ExecuteToString(ctx)
}

func toYaml(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	var err error
	var byts []byte
	if in.IsError() {
		return in
	}
	simpleType := in.ToGoSimpleType(true)
	if byts, err = yaml.MarshalWithOptions(simpleType, yaml.AutoInt()); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}
	v := string(byts)
	v = strings.TrimSuffix(v, "\n")
	return exec.AsValue(v)
}

func nowTime(format string) string {
	return time.Now().Format("20060102150405")
}

func init() {
	gonja.DefaultConfig.VariableStartString = "${"
	gonja.DefaultConfig.VariableEndString = "}"
	gonja.DefaultConfig.TrimBlocks = true

	gonja.DefaultContext.Update(exec.NewContext(map[string]any{"now": nowTime}))
	gonja.DefaultEnvironment.Filters.Register("to_yaml", toYaml)
}
