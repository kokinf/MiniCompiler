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
	errors         []string
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
		errors:         make([]string, 0),
	}

	gen.labelManager = NewLabelManager()
	gen.exprGenerator = NewExpressionGenerator(gen, gen.labelManager)
	gen.cfGenerator = NewControlFlowGenerator(gen, gen.labelManager, gen.exprGenerator)
	gen.arrayGen = NewArrayGenerator(gen, gen.labelManager)
	gen.externGen = NewExternalCallGenerator(gen, gen.labelManager, gen.abi, stdlibParser)
	gen.asmOpt = NewAssemblyOptimizer()

	return gen
}

func (g *X86Generator) addError(msg string) {
	g.errors = append(g.errors, msg)
}

func (g *X86Generator) HasErrors() bool {
	return len(g.errors) > 0
}

func (g *X86Generator) GetErrors() []string {
	return g.errors
}

func (g *X86Generator) Generate() string {
	g.output.Reset()
	g.labelManager.Reset()
	g.stringCounter = 0
	g.errors = nil

	g.emit("; ═══════════════════════════════════════════")
	g.emit("; MiniCompiler x86-64 Code Generator")
	g.emit("; System V AMD64 ABI")
	g.emit("; Sprint 8: Final Release")
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

	if len(g.errors) > 0 {
		return ""
	}

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
	// Структура для хранения информации о переменной
	type VarInfo struct {
		name    string
		size    int
		isLocal bool // true для varToAddr, false для tempToAddr
	}

	// Собираем ВСЕ переменные в правильном порядке
	var vars8 []VarInfo // 8-байтовые
	var vars4 []VarInfo // 4-байтовые

	// Собираем ALLOCA
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Opcode == ir.OpAlloca && inst.Dest != nil {
				name := inst.Dest.Name
				if _, exists := g.tempToAddr[name]; exists {
					continue
				}
				size := 4
				if inst.Src1 != nil {
					if s, ok := inst.Src1.GetIntValue(); ok && s > 0 {
						size = s
					}
				}
				if size == 8 {
					vars8 = append(vars8, VarInfo{name, 8, false})
				} else {
					vars4 = append(vars4, VarInfo{name, size, false})
				}
			}
		}
	}

	// Собираем результаты malloc (8 байт)
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Opcode == ir.OpCall && inst.Dest != nil && inst.Src1 != nil && inst.Src1.Name == "malloc" {
				name := inst.Dest.Name
				if _, exists := g.tempToAddr[name]; !exists {
					vars8 = append(vars8, VarInfo{name, 8, false})
				}
			}
		}
	}

	// Собираем результаты GEP (8 байт)
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Opcode == ir.OpGep && inst.Dest != nil {
				name := inst.Dest.Name
				if _, exists := g.tempToAddr[name]; !exists {
					vars8 = append(vars8, VarInfo{name, 8, false})
				}
			}
		}
	}

	// Собираем Locals
	for varName, varInfo := range fn.Locals {
		if _, exists := g.varToAddr[varName]; exists {
			continue
		}
		size := varInfo.Size
		if size == 0 {
			size = 4
		}
		if size == 8 {
			vars8 = append(vars8, VarInfo{varName, 8, true})
		} else {
			vars4 = append(vars4, VarInfo{varName, size, true})
		}
	}

	// Собираем остальные временные переменные (4 байта)
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Dest != nil && inst.Opcode != ir.OpAlloca && inst.Opcode != ir.OpGep {
				name := inst.Dest.Name
				if _, exists := g.tempToAddr[name]; !exists {
					if _, exists := g.varToAddr[name]; !exists {
						// Проверяем, не 8-байтовая ли это переменная
						is8byte := false
						for _, v := range vars8 {
							if v.name == name {
								is8byte = true
								break
							}
						}
						if !is8byte {
							vars4 = append(vars4, VarInfo{name, 4, false})
						}
					}
				}
			}
		}
	}

	// Выделяем 8-байтовые переменные
	for _, v := range vars8 {
		loc := g.stackFrame.AllocateVar(v.name, v.size)
		addr := fmt.Sprintf("rbp - %d", -loc.StackOffset)
		if v.isLocal {
			g.varToAddr[v.name] = addr
		} else {
			g.tempToAddr[v.name] = addr
		}
		g.tempToAddrSize[v.name] = v.size
	}

	// Выделяем 4-байтовые переменные
	for _, v := range vars4 {
		loc := g.stackFrame.AllocateVar(v.name, v.size)
		addr := fmt.Sprintf("rbp - %d", -loc.StackOffset)
		if v.isLocal {
			g.varToAddr[v.name] = addr
		} else {
			g.tempToAddr[v.name] = addr
		}
		g.tempToAddrSize[v.name] = v.size
	}

	// Padding если нет вызовов функций
	if g.stackFrame.MaxOffset == 0 && g.hasFunctionCalls(fn) {
		g.stackFrame.AllocateVar("_padding", 8)
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
