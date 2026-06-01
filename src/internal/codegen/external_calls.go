package codegen

import (
	"fmt"
	"mikrocompiler/src/internal/ir"
	"mikrocompiler/src/libc"
)

// ExternalCallGenerator генерирует вызовы внешних функций
type ExternalCallGenerator struct {
	gen          *X86Generator
	labelManager *LabelManager
	abi          *ABIInfo
	stdlib       *libc.StdlibParser
}

// NewExternalCallGenerator создает новый генератор внешних вызовов
func NewExternalCallGenerator(gen *X86Generator, labelManager *LabelManager, abi *ABIInfo, stdlib *libc.StdlibParser) *ExternalCallGenerator {
	return &ExternalCallGenerator{
		gen:          gen,
		labelManager: labelManager,
		abi:          abi,
		stdlib:       stdlib,
	}
}

// IsExternFunction проверяет, является ли функция внешней
func (ec *ExternalCallGenerator) IsExternFunction(name string) bool {
	return ec.stdlib.HasFunction(name)
}

// GenerateExternCall генерирует вызов внешней функции
func (ec *ExternalCallGenerator) GenerateExternCall(inst *ir.Instruction) {
	funcName := inst.Src1.Name

	fnDef := ec.stdlib.GetFunction(funcName)
	if fnDef == nil {
		ec.generateGenericExternCall(inst)
		return
	}

	ec.saveCallerSavedRegs()

	// Собираем аргументы: сначала из inst.Args, затем из pendingParams
	args := make([]*ir.Operand, 0)

	// Копируем аргументы из инструкции
	for _, arg := range inst.Args {
		args = append(args, arg)
	}

	// Если аргументов нет, но есть pendingParams - используем их
	if len(args) == 0 && len(ec.gen.pendingParams) > 0 {
		for _, p := range ec.gen.pendingParams {
			if p.Src2 != nil {
				args = append(args, p.Src2)
			}
		}
	}

	// Загружаем аргументы в регистры по ABI
	for i, arg := range args {
		if i < len(ec.abi.IntParamRegs) {
			reg := string(ec.abi.IntParamRegs[i])
			ec.gen.loadToRegister(reg, arg)
		} else {
			// Аргументы сверх 6 на стек
			if arg.Type == ir.OperandLiteral {
				if val, ok := arg.GetIntValue(); ok {
					ec.gen.emit("    push %d", val)
				}
			} else {
				ec.gen.emit("    push %s", arg.Name)
			}
		}
	}

	// Для variadic функций: AL = 0
	if fnDef.IsVariadic {
		ec.gen.emit("    xor eax, eax  ; variadic: no float args")
	}

	// Выравнивание стека
	ec.alignStack()

	// Вызов функции
	ec.gen.emit("    call %s", funcName)

	// Восстановление стека
	ec.restoreStack()

	// Очистка стековых аргументов
	if len(args) > 6 {
		stackArgs := len(args) - 6
		ec.gen.emit("    add rsp, %d  ; cleanup stack args", stackArgs*8)
	}

	// Сохраняем результат
	if inst.Dest != nil {
		switch fnDef.ReturnType {
		case "float", "double":
			ec.gen.emit("    ; float return in xmm0 (not stored)")
		case "void*", "string":
			ec.gen.storeFromRegister("rax", inst.Dest)
		default:
			ec.gen.storeFromRegister("eax", inst.Dest)
		}
	}

	// Проверка malloc на NULL
	if funcName == "malloc" && inst.Dest != nil {
		ec.generateMallocNullCheck()
	}

	ec.restoreCallerSavedRegs()
}

// generateGenericExternCall общий вызов для неизвестных функций
func (ec *ExternalCallGenerator) generateGenericExternCall(inst *ir.Instruction) {
	funcName := inst.Src1.Name

	// Собираем аргументы
	args := make([]*ir.Operand, 0)

	for _, arg := range inst.Args {
		args = append(args, arg)
	}

	if len(args) == 0 && len(ec.gen.pendingParams) > 0 {
		for _, p := range ec.gen.pendingParams {
			if p.Src2 != nil {
				args = append(args, p.Src2)
			}
		}
	}

	ec.saveCallerSavedRegs()

	for i, arg := range args {
		if i < len(ec.abi.IntParamRegs) {
			reg := string(ec.abi.IntParamRegs[i])
			ec.gen.loadToRegister(reg, arg)
		}
	}

	ec.gen.emit("    xor eax, eax")
	ec.alignStack()
	ec.gen.emit("    call %s", funcName)
	ec.restoreStack()

	if inst.Dest != nil {
		ec.gen.storeFromRegister("eax", inst.Dest)
	}

	ec.restoreCallerSavedRegs()
}

// generateMallocNullCheck генерирует проверку malloc на NULL
func (ec *ExternalCallGenerator) generateMallocNullCheck() {
	failLabel := ec.labelManager.NewLabel("malloc_fail")
	okLabel := ec.labelManager.NewLabel("malloc_ok")

	ec.gen.emit("    test rax, rax       ; check malloc result")
	ec.gen.emit("    jz %s               ; NULL → abort", failLabel)
	ec.gen.emit("    jmp %s               ; OK → continue", okLabel)
	ec.gen.emit("%s:", failLabel)
	ec.gen.emit("    call abort           ; malloc failed")
	ec.gen.emit("%s:", okLabel)
}

// saveCallerSavedRegs сохраняет caller-saved регистры перед вызовом
func (ec *ExternalCallGenerator) saveCallerSavedRegs() {
	ec.gen.emit("    ; save caller-saved registers")
	for _, reg := range ec.abi.CallerSavedRegs {
		if reg != RAX {
			ec.gen.emit("    push %s", reg)
		}
	}
}

// restoreCallerSavedRegs восстанавливает caller-saved регистры
func (ec *ExternalCallGenerator) restoreCallerSavedRegs() {
	ec.gen.emit("    ; restore caller-saved registers")
	for i := len(ec.abi.CallerSavedRegs) - 1; i >= 0; i-- {
		reg := ec.abi.CallerSavedRegs[i]
		if reg != RAX {
			ec.gen.emit("    pop %s", reg)
		}
	}
}

// alignStack выравнивает стек на 16 байт
func (ec *ExternalCallGenerator) alignStack() {
	ec.gen.emit("    ; ensure 16-byte stack alignment")
	ec.gen.emit("    push rbx")
	ec.gen.emit("    mov rbx, rsp")
	ec.gen.emit("    and rsp, -16")
}

// restoreStack восстанавливает стек после выравнивания
func (ec *ExternalCallGenerator) restoreStack() {
	ec.gen.emit("    mov rsp, rbx")
	ec.gen.emit("    pop rbx")
}

// GetExternDeclarations возвращает строки extern для ассемблера
func (ec *ExternalCallGenerator) GetExternDeclarations() []string {
	return ec.stdlib.GetExternDeclarations()
}

func (ec *ExternalCallGenerator) String() string {
	return fmt.Sprintf("ExternalCallGenerator(%d stdlib functions)", len(ec.stdlib.GetAllFunctions()))
}
