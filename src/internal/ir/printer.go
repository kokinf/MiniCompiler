package ir

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TextPrinter выводит IR в текстовом формате
type TextPrinter struct {
	showComments bool
}

func NewTextPrinter() *TextPrinter {
	return &TextPrinter{showComments: true}
}

func (p *TextPrinter) Print(program *Program) string {
	return program.String()
}

// DOTPrinter генерирует Graphviz DOT для CFG
type DOTPrinter struct {
	nodeCounter int
}

func NewDOTPrinter() *DOTPrinter {
	return &DOTPrinter{}
}

func (p *DOTPrinter) Print(program *Program) string {
	var sb strings.Builder

	sb.WriteString("digraph CFG {\n")
	sb.WriteString("  node [shape=box, fontname=\"Courier\"];\n")
	sb.WriteString("  edge [arrowhead=vee];\n\n")

	for _, fn := range program.Functions {
		if !fn.IsExtern {
			p.printFunction(&sb, fn)
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (p *DOTPrinter) printFunction(sb *strings.Builder, fn *Function) {
	fmt.Fprintf(sb, "  subgraph cluster_%s {\n", fn.Name)
	fmt.Fprintf(sb, "    label=\"%s\";\n", fn.Name)
	fmt.Fprintf(sb, "    color=blue;\n\n")

	for _, block := range fn.Blocks {
		p.printBlock(sb, block)
	}

	for _, block := range fn.Blocks {
		for _, succ := range block.Successors {
			fmt.Fprintf(sb, "    %s -> %s;\n", block.Label, succ.Label)
		}
	}

	sb.WriteString("  }\n\n")
}

func (p *DOTPrinter) printBlock(sb *strings.Builder, block *BasicBlock) {
	label := block.Label + "\\n"
	for _, inst := range block.Instructions {
		label += strings.ReplaceAll(inst.String(), "\"", "\\\"") + "\\l"
	}

	fmt.Fprintf(sb, "    %s [label=\"%s\"];\n", block.Label, label)
}

// JSONPrinter выводит IR в JSON формате
type JSONPrinter struct{}

func NewJSONPrinter() *JSONPrinter {
	return &JSONPrinter{}
}

type jsonProgram struct {
	Functions []jsonFunction        `json:"functions"`
	Globals   map[string]jsonGlobal `json:"globals,omitempty"`
}

type jsonFunction struct {
	Name       string               `json:"name"`
	ReturnType string               `json:"returnType"`
	IsExtern   bool                 `json:"isExtern,omitempty"`
	Params     []jsonParam          `json:"params"`
	Blocks     []jsonBlock          `json:"blocks,omitempty"`
	Locals     map[string]jsonLocal `json:"locals,omitempty"`
}

type jsonParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type jsonLocal struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Size   int    `json:"size"`
}

type jsonBlock struct {
	Label        string   `json:"label"`
	Instructions []string `json:"instructions"`
	Predecessors []string `json:"predecessors,omitempty"`
	Successors   []string `json:"successors,omitempty"`
}

type jsonGlobal struct {
	Type string      `json:"type"`
	Init interface{} `json:"init,omitempty"`
}

func (p *JSONPrinter) Print(program *Program) string {
	jp := jsonProgram{
		Functions: make([]jsonFunction, 0),
		Globals:   make(map[string]jsonGlobal),
	}

	for name, g := range program.Globals {
		var init interface{}
		if g.Init != nil {
			init = g.Init.Value
		}
		jp.Globals[name] = jsonGlobal{
			Type: g.Type,
			Init: init,
		}
	}

	for _, fn := range program.Functions {
		jf := jsonFunction{
			Name:       fn.Name,
			ReturnType: fn.ReturnType,
			IsExtern:   fn.IsExtern,
			Params:     make([]jsonParam, 0),
			Blocks:     make([]jsonBlock, 0),
			Locals:     make(map[string]jsonLocal),
		}

		for _, p := range fn.Params {
			jf.Params = append(jf.Params, jsonParam{Name: p.Name, Type: p.Type})
		}

		for name, info := range fn.Locals {
			jf.Locals[name] = jsonLocal{
				Type:   info.Type,
				Offset: info.Offset,
				Size:   info.Size,
			}
		}

		for _, block := range fn.Blocks {
			jb := jsonBlock{
				Label:        block.Label,
				Instructions: make([]string, 0),
				Predecessors: make([]string, 0),
				Successors:   make([]string, 0),
			}

			for _, inst := range block.Instructions {
				jb.Instructions = append(jb.Instructions, inst.String())
			}

			for _, pred := range block.Predecessors {
				jb.Predecessors = append(jb.Predecessors, pred.Label)
			}

			for _, succ := range block.Successors {
				jb.Successors = append(jb.Successors, succ.Label)
			}

			jf.Blocks = append(jf.Blocks, jb)
		}

		jp.Functions = append(jp.Functions, jf)
	}

	bytes, _ := json.MarshalIndent(jp, "", "  ")
	return string(bytes)
}

// Stats собирает статистику по IR
type Stats struct {
	InstructionCounts map[string]int
	TotalInstructions int
	BasicBlockCount   int
	TempCount         int
	FunctionCount     int
	PhiNodeCount      int
	DefUseChains      int
}

func CollectStats(program *Program) *Stats {
	s := &Stats{
		InstructionCounts: make(map[string]int),
	}

	s.FunctionCount = len(program.Functions)

	for _, fn := range program.Functions {
		s.BasicBlockCount += len(fn.Blocks)
		s.TempCount += fn.tempCounter

		for _, block := range fn.Blocks {
			for _, inst := range block.Instructions {
				s.TotalInstructions++
				s.InstructionCounts[inst.Opcode.String()]++

				if inst.Opcode == OpPhi {
					s.PhiNodeCount++
				}
			}
		}

		s.DefUseChains += len(fn.GlobalDefs)
	}

	return s
}

func (s *Stats) String() string {
	var sb strings.Builder

	sb.WriteString("IR Statistics:\n")
	sb.WriteString("================\n\n")

	fmt.Fprintf(&sb, "Functions:       %d\n", s.FunctionCount)
	fmt.Fprintf(&sb, "Basic Blocks:    %d\n", s.BasicBlockCount)
	fmt.Fprintf(&sb, "Instructions:    %d\n", s.TotalInstructions)
	fmt.Fprintf(&sb, "Temporaries:     %d\n", s.TempCount)
	fmt.Fprintf(&sb, "Phi Nodes:       %d\n", s.PhiNodeCount)
	fmt.Fprintf(&sb, "Def-Use Chains:  %d\n\n", s.DefUseChains)

	sb.WriteString("Instruction breakdown:\n")
	for op, count := range s.InstructionCounts {
		fmt.Fprintf(&sb, "  %-12s: %d\n", op, count)
	}

	return sb.String()
}
