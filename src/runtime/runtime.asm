; MiniCompiler Runtime Library for x86-64 Linux
; System V AMD64 ABI

section .data
    buffer db 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
    newline db 10
    negative_sign db '-'

section .text

; ============================================================================
; print_int: prints integer from rdi to stdout
; Input: rdi = integer to print
; ============================================================================
global print_int
print_int:
    push rbp
    mov rbp, rsp
    push rbx
    push r12
    
    mov r12, rdi        ; save the number
    
    ; Handle zero
    test r12, r12
    jnz .not_zero
    mov byte [buffer], '0'
    mov byte [buffer+1], 10
    mov rax, 1
    mov rdi, 1
    mov rsi, buffer
    mov rdx, 2
    syscall
    jmp .done
    
.not_zero:
    ; Handle negative numbers
    mov rbx, 0          ; flag for negative
    cmp r12, 0
    jge .convert
    mov rbx, 1
    neg r12
    
.convert:
    ; Convert integer to string (in reverse)
    mov rcx, buffer + 20
    mov byte [rcx], 0
    dec rcx
    
.convert_loop:
    mov rax, r12
    mov rdx, 0
    mov rsi, 10
    div rsi             ; rax = quotient, rdx = remainder
    mov r12, rax        ; save quotient
    add dl, '0'
    mov [rcx], dl
    dec rcx
    test r12, r12
    jnz .convert_loop
    
    ; Add negative sign if needed
    test rbx, rbx
    jz .write
    mov byte [rcx], '-'
    dec rcx
    
.write:
    inc rcx
    
    ; Calculate length
    mov rdx, buffer + 20
    sub rdx, rcx
    
    ; Write to stdout
    mov rax, 1          ; syscall: write
    mov rdi, 1          ; fd: stdout
    mov rsi, rcx        ; buffer
    ; rdx already contains count
    syscall
    
.done:
    pop r12
    pop rbx
    pop rbp
    ret

; ============================================================================
; print_string: prints null-terminated string to stdout
; Input: rdi = pointer to string
; ============================================================================
global print_string
print_string:
    push rbp
    mov rbp, rsp
    
    ; Calculate string length
    mov rsi, rdi
    xor rdx, rdx
.strlen:
    cmp byte [rsi + rdx], 0
    je .write
    inc rdx
    jmp .strlen
    
.write:
    ; Write to stdout
    mov rax, 1          ; syscall: write
    mov rdi, 1          ; fd: stdout
    ; rsi already points to string
    ; rdx already contains length
    syscall
    
    pop rbp
    ret

; ============================================================================
; print_newline: prints newline character
; ============================================================================
global print_newline
print_newline:
    push rbp
    mov rbp, rsp
    
    mov rax, 1          ; syscall: write
    mov rdi, 1          ; fd: stdout
    mov rsi, newline
    mov rdx, 1
    syscall
    
    pop rbp
    ret

; ============================================================================
; read_int: reads integer from stdin, returns in rax
; ============================================================================
global read_int
read_int:
    push rbp
    mov rbp, rsp
    push rbx
    
    ; Read from stdin
    mov rax, 0          ; syscall: read
    mov rdi, 0          ; fd: stdin
    mov rsi, buffer
    mov rdx, 20
    syscall
    
    ; Convert string to integer
    xor rax, rax
    xor rcx, rcx
    mov rsi, buffer
    xor rbx, rbx        ; negative flag
    
    ; Check for negative sign
    cmp byte [rsi], '-'
    jne .atoi_loop
    inc rsi
    mov rbx, 1
    
.atoi_loop:
    movzx rdx, byte [rsi]
    cmp dl, 10          ; newline
    je .done
    cmp dl, 0
    je .done
    cmp dl, '0'
    jl .done
    cmp dl, '9'
    jg .done
    
    sub dl, '0'
    imul rax, 10
    add rax, rdx
    
    inc rsi
    jmp .atoi_loop
    
.done:
    ; Apply sign
    test rbx, rbx
    jz .exit
    neg rax
    
.exit:
    pop rbx
    pop rbp
    ret

; ============================================================================
; exit: exit program with status code in rdi
; ============================================================================
global exit
exit:
    mov rax, 60         ; syscall: exit
    syscall
    ; Does not return

; ============================================================================
; _start: program entry point
; NOTE: This is defined in the generated assembly code, not here
; ============================================================================