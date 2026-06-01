// src/internal/codegen/array_generator.go
package codegen

import (
	"fmt"
	"mikrocompiler/src/internal/ir"
)

type ArrayGenerator struct {
	gen          *X86Generator
	labelManager *LabelManager
}

func NewArrayGenerator(gen *X86Generator, labelManager *LabelManager) *ArrayGenerator {
	return &ArrayGenerator{
		gen:          gen,
		labelManager: labelManager,
	}
}

// GenerateMalloc генерирует вызов malloc с проверкой на NULL
func (ag *ArrayGenerator) GenerateMalloc(inst *ir.Instruction) {
	if inst.Src1 == nil || inst.Dest == nil {
		return
	}

	ag.gen.loadToRegister("edi", inst.Src1)

	ag.gen.emit("    push rbx")
	ag.gen.emit("    mov rbx, rsp")
	ag.gen.emit("    and rsp, -16")
	ag.gen.emit("    call malloc                             ; %s = malloc(%s)", inst.Dest, inst.Src1)
	ag.gen.emit("    mov rsp, rbx")
	ag.gen.emit("    pop rbx")

	ag.gen.emit("    test rax, rax                           ; проверка результата malloc на NULL")
	if inst.Dest != nil {
		ag.gen.storeFromRegister("rax", inst.Dest)
	}

	failLabel := ag.labelManager.NewLabel("malloc_fail")
	okLabel := ag.labelManager.NewLabel("malloc_ok")

	ag.gen.emit("    jz %s                                   ; переход если NULL -> abort", failLabel)
	ag.gen.emit("    jmp %s                                   ; OK -> продолжаем", okLabel)
	ag.gen.emit("%s:", failLabel)
	ag.gen.emit("    call abort                               ; malloc вернул NULL - аварийное завершение")
	ag.gen.emit("%s:", okLabel)
}

// GenerateFree генерирует вызов free для указателя
func (ag *ArrayGenerator) GenerateFree(ptrOp *ir.Operand) {
	if ptrOp == nil {
		return
	}
	ag.gen.loadToRegister("rdi", ptrOp)
	ag.gen.emit("    call free                               ; free(%s)", ptrOp.Name)
}

// GenerateGEP генерирует вычисление адреса элемента массива
func (ag *ArrayGenerator) GenerateGEP(inst *ir.Instruction, elementSize int) {
	if inst.Src1 == nil || inst.Src2 == nil || inst.Dest == nil {
		return
	}

	ag.gen.loadToRegister("rax", inst.Src1)

	if inst.Src2.Type == ir.OperandLiteral {
		if val, ok := inst.Src2.GetIntValue(); ok {
			ag.gen.emit("    mov rcx, %d                             ; индекс = %d", val, val)
		} else {
			ag.gen.emit("    xor ecx, ecx                            ; индекс = 0")
		}
	} else {
		ag.gen.emit("    movsxd rcx, dword [%s]                   ; загружаем индекс", ag.gen.getOperandAddr(inst.Src2))
	}

	switch elementSize {
	case 1:
		// размер 1 - сдвиг не нужен
	case 2:
		ag.gen.emit("    shl rcx, 1                              ; индекс * 2 (sizeof(int16))")
	case 4:
		ag.gen.emit("    shl rcx, 2                              ; индекс * 4 (sizeof(int32))")
	case 8:
		ag.gen.emit("    shl rcx, 3                              ; индекс * 8 (sizeof(int64))")
	default:
		ag.gen.emit("    imul rcx, %d                             ; индекс * %d (размер элемента)", elementSize, elementSize)
	}

	ag.gen.emit("    add rax, rcx                            ; %s = GEP %s, %s (адрес элемента массива)", inst.Dest, inst.Src1, inst.Src2)
	ag.gen.storeFromRegister("rax", inst.Dest)

	if inst.Dest != nil {
		ag.gen.tempToAddrSize[inst.Dest.Name] = 8
	}
}

// GenerateArrayLoad генерирует загрузку элемента массива
func (ag *ArrayGenerator) GenerateArrayLoad(inst *ir.Instruction, elementSize int) {
	if inst.Src1 == nil || inst.Dest == nil {
		return
	}

	addr := ag.gen.getOperandAddr(inst.Src1)
	ag.gen.emit("    mov rax, [%s]                           ; загружаем адрес элемента", addr)

	switch elementSize {
	case 1:
		ag.gen.emit("    movzx eax, byte [rax]                   ; %s = LOAD [%s] (загружаем byte)", inst.Dest, inst.Src1)
	case 2:
		ag.gen.emit("    movzx eax, word [rax]                   ; %s = LOAD [%s] (загружаем word)", inst.Dest, inst.Src1)
	case 4:
		ag.gen.emit("    mov eax, [rax]                          ; %s = LOAD [%s] (загружаем dword)", inst.Dest, inst.Src1)
	case 8:
		ag.gen.emit("    mov rax, [rax]                          ; %s = LOAD [%s] (загружаем qword)", inst.Dest, inst.Src1)
	}

	ag.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateArrayStore генерирует сохранение элемента массива
func (ag *ArrayGenerator) GenerateArrayStore(inst *ir.Instruction, elementSize int) {
	if inst.Src1 == nil || inst.Src2 == nil {
		return
	}

	addr := ag.gen.getOperandAddr(inst.Src1)
	ag.gen.emit("    mov rcx, [%s]                           ; загружаем адрес для сохранения", addr)
	ag.gen.loadToRegister("eax", inst.Src2)

	switch elementSize {
	case 1:
		ag.gen.emit("    mov [rcx], al                           ; STORE [%s], %s (сохраняем byte)", inst.Src1, inst.Src2)
	case 2:
		ag.gen.emit("    mov [rcx], ax                           ; STORE [%s], %s (сохраняем word)", inst.Src1, inst.Src2)
	case 4:
		ag.gen.emit("    mov [rcx], eax                          ; STORE [%s], %s (сохраняем dword)", inst.Src1, inst.Src2)
	case 8:
		ag.gen.emit("    mov [rcx], rax                          ; STORE [%s], %s (сохраняем qword)", inst.Src1, inst.Src2)
	}
}

// GenerateArrayInit генерирует инициализацию массива значениями
func (ag *ArrayGenerator) GenerateArrayInit(basePtr string, values []int, elementSize int) {
	for i, val := range values {
		offset := i * elementSize
		if elementSize == 4 {
			ag.gen.emit("    mov dword [%s + %d], %d                  ; arr[%d] = %d", basePtr, offset, val, i, val)
		} else if elementSize == 8 {
			ag.gen.emit("    mov qword [%s + %d], %d                  ; arr[%d] = %d", basePtr, offset, val, i, val)
		}
	}
}

// GenerateArrayCopy генерирует копирование массива через memcpy
func (ag *ArrayGenerator) GenerateArrayCopy(dest, src string, size int) {
	if size <= 32 {
		for i := 0; i < size; i += 4 {
			ag.gen.emit("    mov eax, [%s + %d]                      ; копируем элемент %d", src, i, i/4)
			ag.gen.emit("    mov [%s + %d], eax", dest, i)
		}
	} else {
		ag.gen.emit("    mov rdi, %s                             ; dest = %s", dest, dest)
		ag.gen.emit("    mov rsi, %s                             ; src = %s", src, src)
		ag.gen.emit("    mov rdx, %d                             ; size = %d", size, size)
		ag.gen.emit("    call memcpy                              ; копируем массив")
	}
}

func (ag *ArrayGenerator) String() string {
	return fmt.Sprintf("ArrayGenerator")
}
