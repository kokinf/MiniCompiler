package libc

import (
	"os"
	"strings"
)

type StdlibFunction struct {
	ReturnType string
	Name       string
	Parameters []StdlibParam
	IsVariadic bool
}

type StdlibParam struct {
	Type string
	Name string
}

type StdlibParser struct {
	functions map[string]*StdlibFunction
}

func NewStdlibParser() *StdlibParser {
	return &StdlibParser{
		functions: make(map[string]*StdlibFunction),
	}
}

func (sp *StdlibParser) ParseFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return sp.Parse(string(content))
}

func (sp *StdlibParser) Parse(content string) error {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "//") || line == "" {
			continue
		}
		if !strings.HasPrefix(line, "extern ") {
			continue
		}

		line = strings.TrimPrefix(line, "extern ")
		line = strings.TrimSuffix(line, ";")
		line = strings.TrimSpace(line)

		funcDef := sp.parseFunctionDef(line)
		if funcDef != nil {
			sp.functions[funcDef.Name] = funcDef
		}
	}

	return nil
}

func (sp *StdlibParser) parseFunctionDef(def string) *StdlibFunction {
	parenIdx := strings.Index(def, "(")
	if parenIdx == -1 {
		return nil
	}

	beforeParen := strings.TrimSpace(def[:parenIdx])
	parts := strings.Fields(beforeParen)
	if len(parts) < 2 {
		return nil
	}

	returnType := strings.Join(parts[:len(parts)-1], " ")
	funcName := parts[len(parts)-1]

	paramsStr := def[parenIdx+1:]
	paramsStr = strings.TrimSuffix(paramsStr, ")")
	paramsStr = strings.TrimSpace(paramsStr)

	funcDef := &StdlibFunction{
		ReturnType: returnType,
		Name:       funcName,
		Parameters: make([]StdlibParam, 0),
		IsVariadic: false,
	}

	if paramsStr == "" || paramsStr == "void" {
		return funcDef
	}

	if strings.Contains(paramsStr, "...") {
		funcDef.IsVariadic = true
		paramsStr = strings.Replace(paramsStr, "...", "", -1)
		paramsStr = strings.TrimSuffix(paramsStr, ",")
		paramsStr = strings.TrimSpace(paramsStr)
	}

	if paramsStr == "" {
		return funcDef
	}

	paramParts := strings.Split(paramsStr, ",")
	for _, paramPart := range paramParts {
		paramPart = strings.TrimSpace(paramPart)
		if paramPart == "" {
			continue
		}
		param := sp.parseParam(paramPart)
		if param.Type != "" {
			funcDef.Parameters = append(funcDef.Parameters, param)
		}
	}

	return funcDef
}

func (sp *StdlibParser) parseParam(paramStr string) StdlibParam {
	parts := strings.Fields(paramStr)
	if len(parts) < 2 {
		if len(parts) == 1 {
			return StdlibParam{Type: parts[0], Name: ""}
		}
		return StdlibParam{}
	}

	name := parts[len(parts)-1]
	typ := strings.Join(parts[:len(parts)-1], " ")

	return StdlibParam{
		Type: typ,
		Name: name,
	}
}

func (sp *StdlibParser) GetFunction(name string) *StdlibFunction {
	return sp.functions[name]
}

func (sp *StdlibParser) GetAllFunctions() map[string]*StdlibFunction {
	return sp.functions
}

func (sp *StdlibParser) HasFunction(name string) bool {
	_, exists := sp.functions[name]
	return exists
}

func (sp *StdlibParser) GetExternDeclarations() []string {
	decls := make([]string, 0, len(sp.functions))
	for name := range sp.functions {
		decls = append(decls, "extern "+name)
	}
	return decls
}

func (sp *StdlibParser) GetFunctionNames() []string {
	names := make([]string, 0, len(sp.functions))
	for name := range sp.functions {
		names = append(names, name)
	}
	return names
}
