package codegen

import (
	"mikrocompiler/src/internal/semantic"
)

// Register представляет x86-64 регистр
type Register string

const (
	RAX Register = "rax"
	RBX Register = "rbx"
	RCX Register = "rcx"
	RDX Register = "rdx"
	RSI Register = "rsi"
	RDI Register = "rdi"
	RBP Register = "rbp"
	RSP Register = "rsp"
	R8  Register = "r8"
	R9  Register = "r9"
	R10 Register = "r10"
	R11 Register = "r11"
	R12 Register = "r12"
	R13 Register = "r13"
	R14 Register = "r14"
	R15 Register = "r15"

	EAX Register = "eax"
	EBX Register = "ebx"
	ECX Register = "ecx"
	EDX Register = "edx"
	ESI Register = "esi"
	EDI Register = "edi"
	EBP Register = "ebp"
	ESP Register = "esp"

	XMM0 Register = "xmm0"
	XMM1 Register = "xmm1"
	XMM2 Register = "xmm2"
	XMM3 Register = "xmm3"
	XMM4 Register = "xmm4"
	XMM5 Register = "xmm5"
	XMM6 Register = "xmm6"
	XMM7 Register = "xmm7"
)

type OperandSize string

const (
	SizeByte  OperandSize = "byte"
	SizeWord  OperandSize = "word"
	SizeDword OperandSize = "dword"
	SizeQword OperandSize = "qword"
)

type VarLocation struct {
	IsRegister  bool
	Register    Register
	StackOffset int
	Size        int
}

type StackFrame struct {
	LocalVars     map[string]*VarLocation
	CurrentOffset int
	MaxOffset     int
}

func NewStackFrame() *StackFrame {
	return &StackFrame{
		LocalVars: make(map[string]*VarLocation),
	}
}

func (sf *StackFrame) AllocateVar(name string, size int) *VarLocation {
	// Выравнивание на размер переменной (но не более 8 байт)
	align := size
	if align > 8 {
		align = 8
	}

	// Выравниваем текущий offset
	remainder := sf.CurrentOffset % align
	if remainder != 0 {
		sf.CurrentOffset += align - remainder
	}

	sf.CurrentOffset += size

	loc := &VarLocation{
		IsRegister:  false,
		StackOffset: -sf.CurrentOffset,
		Size:        size,
	}
	sf.LocalVars[name] = loc

	if sf.CurrentOffset > sf.MaxOffset {
		sf.MaxOffset = sf.CurrentOffset
	}

	return loc
}

func (sf *StackFrame) GetStackSize() int {
	size := sf.MaxOffset
	// Выравниваем размер стека на 16 байт (требование ABI)
	if size%16 != 0 {
		size += 16 - (size % 16)
	}
	return size
}

type ABIInfo struct {
	IntParamRegs    []Register
	FloatParamRegs  []Register
	CallerSavedRegs []Register
	CalleeSavedRegs []Register
}

func NewABIInfo() *ABIInfo {
	return &ABIInfo{
		IntParamRegs:    []Register{RDI, RSI, RDX, RCX, R8, R9},
		FloatParamRegs:  []Register{XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6, XMM7},
		CallerSavedRegs: []Register{RAX, RCX, RDX, RSI, RDI, R8, R9, R10, R11},
		CalleeSavedRegs: []Register{RBX, R12, R13, R14, R15, RBP, RSP},
	}
}

func (abi *ABIInfo) GetIntParamReg(index int) (Register, bool) {
	if index < len(abi.IntParamRegs) {
		return abi.IntParamRegs[index], true
	}
	return "", false
}

func (abi *ABIInfo) GetFloatParamReg(index int) (Register, bool) {
	if index < len(abi.FloatParamRegs) {
		return abi.FloatParamRegs[index], true
	}
	return "", false
}

func (abi *ABIInfo) GetStackParamOffset(index int) int {
	return 16 + (index-6)*8
}

type TypeInfo struct {
	Type        *semantic.Type
	Size        int
	Align       int
	IsFloat     bool
	OperandSize OperandSize
}

func GetTypeInfo(t *semantic.Type) *TypeInfo {
	if t == nil {
		return &TypeInfo{Size: 8, Align: 8, OperandSize: SizeQword}
	}

	info := &TypeInfo{Type: t}

	switch t.Kind {
	case semantic.TypeInt:
		info.Size = 4
		info.Align = 4
		info.IsFloat = false
		info.OperandSize = SizeDword
	case semantic.TypeFloat:
		info.Size = 8
		info.Align = 8
		info.IsFloat = true
		info.OperandSize = SizeQword
	case semantic.TypeBool:
		info.Size = 1
		info.Align = 1
		info.IsFloat = false
		info.OperandSize = SizeByte
	case semantic.TypeString:
		info.Size = 16
		info.Align = 8
		info.IsFloat = false
		info.OperandSize = SizeQword
	case semantic.TypeArray, semantic.TypePointer:
		info.Size = 8
		info.Align = 8
		info.IsFloat = false
		info.OperandSize = SizeQword
	default:
		info.Size = 8
		info.Align = 8
		info.IsFloat = false
		info.OperandSize = SizeQword
	}

	return info
}

func GetReg32(reg64 Register) Register {
	mapping := map[Register]Register{
		RAX: EAX, RBX: EBX, RCX: ECX, RDX: EDX,
		RSI: ESI, RDI: EDI, RBP: EBP, RSP: ESP,
	}
	if reg32, ok := mapping[reg64]; ok {
		return reg32
	}
	return reg64
}
