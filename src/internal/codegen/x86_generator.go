// src/internal/codegen/x86_generator.go
package codegen

import (
	"fmt"
	"strings"

	"mikrocompiler/src/internal/ir"
	"mikrocompiler/src/internal/semantic"
	"mikrocompiler/src/libc"
)

type X86Generator struct {
	output         strings.Builder
	program        *ir.Program
	symbolTable    *semantic.SymbolTable
	typeSystem     *semantic.TypeSystem
	abi            *ABIInfo
	currentFunc    *ir.Function
	stackFrame     *StackFrame
	labelManager   *LabelManager
	exprGenerator  *ExpressionGenerator
	cfGenerator    *ControlFlowGenerator
	arrayGen       *ArrayGenerator
	externGen      *ExternalCallGenerator
	asmOpt         *AssemblyOptimizer
	stdlibParser   *libc.StdlibParser
	returnEmitted  bool
	stringCounter  int
	tempToAddrSize map[string]int
	tempToAddr     map[string]string
	varToAddr      map[string]string
	pendingParams  []*ir.Instruction
}

func NewX86Generator(program *ir.Program, symbolTable *semantic.SymbolTable, typeSystem *semantic.TypeSystem) *X86Generator {
	stdlibParser := libc.NewStdlibParser()
	stdlibParser.ParseFile("src/libc/stdlib.h")

	gen := &X86Generator{
		program:        program,
		symbolTable:    symbolTable,
		typeSystem:     typeSystem,
		abi:            NewABIInfo(),
		tempToAddr:     make(map[string]string),
		tempToAddrSize: make(map[string]int),
		varToAddr:      make(map[string]string),
		stdlibParser:   stdlibParser,
	}

	gen.labelManager = NewLabelManager()
	gen.exprGenerator = NewExpressionGenerator(gen, gen.labelManager)
	gen.cfGenerator = NewControlFlowGenerator(gen, gen.labelManager, gen.exprGenerator)
	gen.arrayGen = NewArrayGenerator(gen, gen.labelManager)
	gen.externGen = NewExternalCallGenerator(gen, gen.labelManager, gen.abi, stdlibParser)
	gen.asmOpt = NewAssemblyOptimizer()

	return gen
}

func (g *X86Generator) Generate() string {
	g.output.Reset()
	g.labelManager.Reset()
	g.stringCounter = 0

	g.emit("; ═══════════════════════════════════════════")
	g.emit("; MiniCompiler x86-64 Code Generator")
	g.emit("; System V AMD64 ABI")
	g.emit("; Sprint 7: Arrays, Extern, Optimizations")
	g.emit("; ═══════════════════════════════════════════")
	g.emit("")

	g.emit("; External functions from stdlib.h")
	for _, decl := range g.externGen.GetExternDeclarations() {
		g.emit(decl)
	}
	g.emit("extern print_int, print_string, print_newline, read_int, exit")
	g.emit("")
	g.emit("section .data")
	g.emitStrings()
	g.emit("")
	g.emit("section .text")
	g.emit("")

	for _, fn := range g.program.Functions {
		if !fn.IsExtern {
			g.generateFunction(fn)
		}
	}

	g.generateStart()
	return g.output.String()
}

func (g *X86Generator) emit(format string, args ...interface{}) {
	g.output.WriteString(fmt.Sprintf(format, args...))
	g.output.WriteString("\n")
}

func (g *X86Generator) emitStrings() {
	for _, gvar := range g.program.Globals {
		if gvar.Type == "string" && gvar.Init != nil {
			if str, ok := gvar.Init.Value.(string); ok {
				label := g.labelManager.StringLabel(g.stringCounter)
				g.stringCounter++
				g.emit("%s: db \"%s\", 0", label, str)
			}
		}
	}
}

func (g *X86Generator) generateFunction(fn *ir.Function) {
	g.currentFunc = fn
	g.stackFrame = NewStackFrame()
	g.returnEmitted = false
	g.tempToAddr = make(map[string]string)
	g.tempToAddrSize = make(map[string]int)
	g.varToAddr = make(map[string]string)
	g.pendingParams = nil

	g.emit("")
	g.emit("; Function: %s", fn.Name)
	g.emit("global %s", fn.Name)
	g.emit("%s:", fn.Name)

	g.collectAllocations(fn)
	g.cfGenerator.GenerateFunctionPrologue(fn)
	g.setupParameters(fn)

	for _, block := range fn.Blocks {
		g.cfGenerator.GenerateBasicBlock(block)
	}

	if !g.returnEmitted {
		g.emit("    xor eax, eax")
		g.cfGenerator.GenerateFunctionEpilogue()
	}
}

func (g *X86Generator) collectAllocations(fn *ir.Function) {
	// ALLOCA
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Opcode == ir.OpAlloca && inst.Dest != nil {
				name := inst.Dest.Name
				if _, exists := g.tempToAddr[name]; !exists {
					size := 4
					if inst.Src1 != nil {
						if s, ok := inst.Src1.GetIntValue(); ok && s > 0 {
							size = s
						}
					}
					loc := g.stackFrame.AllocateVar(name, size)
					g.tempToAddr[name] = fmt.Sprintf("rbp - %d", -loc.StackOffset)
					g.tempToAddrSize[name] = size
				}
			}
			// malloc возвращает 8 байт
			if inst.Opcode == ir.OpCall && inst.Dest != nil && inst.Src1 != nil && inst.Src1.Name == "malloc" {
				if _, exists := g.tempToAddr[inst.Dest.Name]; !exists {
					loc := g.stackFrame.AllocateVar(inst.Dest.Name, 8)
					g.tempToAddr[inst.Dest.Name] = fmt.Sprintf("rbp - %d", -loc.StackOffset)
				}
				g.tempToAddrSize[inst.Dest.Name] = 8
			}
			// GEP возвращает 8 байт
			if inst.Opcode == ir.OpGep && inst.Dest != nil {
				if _, exists := g.tempToAddr[inst.Dest.Name]; !exists {
					loc := g.stackFrame.AllocateVar(inst.Dest.Name, 8)
					g.tempToAddr[inst.Dest.Name] = fmt.Sprintf("rbp - %d", -loc.StackOffset)
				}
				g.tempToAddrSize[inst.Dest.Name] = 8
			}
		}
	}

	// Locals
	for varName := range fn.Locals {
		if _, exists := g.varToAddr[varName]; !exists {
			varInfo := fn.Locals[varName]
			size := varInfo.Size
			if size == 0 {
				size = 4
			}
			loc := g.stackFrame.AllocateVar(varName, size)
			g.varToAddr[varName] = fmt.Sprintf("rbp - %d", -loc.StackOffset)
		}
	}

	// Остальные temp
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Dest != nil && inst.Opcode != ir.OpAlloca {
				name := inst.Dest.Name
				if _, exists := g.tempToAddr[name]; !exists {
					if _, exists := g.varToAddr[name]; !exists {
						loc := g.stackFrame.AllocateVar(name, 4)
						g.tempToAddr[name] = fmt.Sprintf("rbp - %d", -loc.StackOffset)
					}
				}
			}
		}
	}
}

func (g *X86Generator) getOperandAddr(op *ir.Operand) string {
	if op == nil {
		return "rbp - 4"
	}

	switch op.Type {
	case ir.OperandVar:
		if addr, ok := g.varToAddr[op.Name]; ok {
			return addr
		}
		if addr, ok := g.tempToAddr[op.Name]; ok {
			return addr
		}
		loc := g.stackFrame.AllocateVar(op.Name, 8)
		addr := fmt.Sprintf("rbp - %d", -loc.StackOffset)
		g.varToAddr[op.Name] = addr
		return addr

	case ir.OperandTemp:
		if addr, ok := g.tempToAddr[op.Name]; ok {
			return addr
		}
		loc := g.stackFrame.AllocateVar(op.Name, 4)
		addr := fmt.Sprintf("rbp - %d", -loc.StackOffset)
		g.tempToAddr[op.Name] = addr
		return addr

	default:
		return fmt.Sprintf("rbp - %d", -(g.stackFrame.MaxOffset + 8))
	}
}

func (g *X86Generator) getOperandSize(op *ir.Operand) int {
	if op == nil {
		return 4
	}
	if op.Name != "" {
		if s, ok := g.tempToAddrSize[op.Name]; ok {
			return s
		}
		if varInfo, ok := g.currentFunc.Locals[op.Name]; ok {
			return varInfo.Size
		}
		for _, param := range g.currentFunc.Params {
			if param.Name == op.Name {
				if param.Type == "int[]" || param.Type == "float[]" ||
					param.Type == "bool[]" || param.Type == "string[]" ||
					param.Type == "pointer" || param.Type == "void*" || param.Type == "string" {
					return 8
				}
				return 4
			}
		}
	}
	return 4
}

func (g *X86Generator) hasFunctionCalls(fn *ir.Function) bool {
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Opcode == ir.OpCall {
				return true
			}
		}
	}
	return false
}

func (g *X86Generator) isExternCall(funcName string) bool {
	return g.stdlibParser.HasFunction(funcName)
}

func (g *X86Generator) setupParameters(fn *ir.Function) {
	if len(fn.Params) == 0 {
		return
	}

	for i, param := range fn.Params {
		if reg, ok := g.abi.GetIntParamReg(i); ok {
			addr := g.getOperandAddr(&ir.Operand{Type: ir.OperandVar, Name: param.Name})
			if param.Type == "int[]" || param.Type == "float[]" ||
				param.Type == "bool[]" || param.Type == "string[]" ||
				param.Type == "pointer" || param.Type == "void*" || param.Type == "string" {
				g.emit("    mov [%s], %s", addr, string(reg))
			} else {
				g.emit("    mov [%s], %s", addr, string(GetReg32(reg)))
			}
		}
	}
}

func (g *X86Generator) generateStart() {
	hasMain := false
	for _, fn := range g.program.Functions {
		if fn.Name == "main" {
			hasMain = true
			break
		}
	}

	if hasMain {
		g.emit("")
		g.emit("; _start is provided by libc (crt1.o) - link with gcc")
	}
}

func (g *X86Generator) loadToRegister(reg string, op *ir.Operand) {
	if op == nil {
		g.emit("    xor %s, %s                              ; обнуляем регистр", reg, reg)
		return
	}

	switch op.Type {
	case ir.OperandLiteral:
		if val, ok := op.GetIntValue(); ok {
			g.emit("    mov %s, %d                               ; загружаем константу %d", reg, val, val)
		} else if val, ok := op.GetBoolValue(); ok {
			if val {
				g.emit("    mov %s, 1                                ; загружаем true", reg)
			} else {
				g.emit("    xor %s, %s                              ; загружаем false", reg, reg)
			}
		} else {
			g.emit("    mov %s, %v", reg, op.Value)
		}

	case ir.OperandVar, ir.OperandTemp:
		addr := g.getOperandAddr(op)
		loadSize := g.getOperandSize(op)
		if loadSize == 8 {
			g.emit("    mov %s, [%s]                            ; загружаем %s (64 бита)", reg, addr, op.Name)
		} else {
			g.emit("    mov %s, [%s]                            ; загружаем %s (32 бита)", reg, addr, op.Name)
		}

	default:
		g.emit("    mov %s, %s", reg, op.Name)
	}
}

func (g *X86Generator) storeFromRegister(reg string, dest *ir.Operand) {
	if dest == nil {
		return
	}

	switch dest.Type {
	case ir.OperandVar, ir.OperandTemp:
		addr := g.getOperandAddr(dest)
		g.emit("    mov [%s], %s                             ; сохраняем %s", addr, reg, dest.Name)
	}
}
