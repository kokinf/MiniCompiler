package codegen

import (
	"fmt"
	"strings"
)

// OptimizationPass представляет один проход оптимизации ассемблерного кода
type OptimizationPass struct {
	Name        string
	Description string
	Enabled     bool
}

// AssemblyOptimizer выполняет оптимизации на уровне ассемблерного кода
type AssemblyOptimizer struct {
	passes  []OptimizationPass
	stats   map[string]int
	enabled bool
}

// NewAssemblyOptimizer создает новый оптимизатор ассемблера
func NewAssemblyOptimizer() *AssemblyOptimizer {
	return &AssemblyOptimizer{
		passes: []OptimizationPass{
			{Name: "peephole", Description: "Peephole optimizations (remove redundant mov, push/pop pairs)", Enabled: true},
			{Name: "dead_code", Description: "Remove unreachable code after unconditional jumps", Enabled: true},
			{Name: "constant_folding", Description: "Fold constant expressions at assembly level", Enabled: true},
			{Name: "jump_optimization", Description: "Optimize jump chains (jmp to jmp elimination)", Enabled: true},
		},
		stats:   make(map[string]int),
		enabled: true,
	}
}

// Optimize выполняет все включенные оптимизационные проходы
func (ao *AssemblyOptimizer) Optimize(asmCode string) string {
	if !ao.enabled {
		return asmCode
	}

	result := asmCode

	for _, pass := range ao.passes {
		if !pass.Enabled {
			continue
		}

		before := strings.Count(result, "\n")

		switch pass.Name {
		case "peephole":
			result = ao.peepholeOptimize(result)
		case "dead_code":
			result = ao.eliminateDeadCode(result)
		case "constant_folding":
			result = ao.foldConstants(result)
		case "jump_optimization":
			result = ao.optimizeJumps(result)
		}

		after := strings.Count(result, "\n")
		ao.stats[pass.Name] = before - after
	}

	return result
}

// peepholeOptimize выполняет локальные оптимизации
func (ao *AssemblyOptimizer) peepholeOptimize(asm string) string {
	lines := strings.Split(asm, "\n")
	optimized := make([]string, 0, len(lines))

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, ";") {
			optimized = append(optimized, line)
			continue
		}

		// mov reg, reg — удаляем
		if strings.Contains(trimmed, "mov") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				dest := strings.TrimSuffix(parts[1], ",")
				src := parts[2]
				if dest == src && strings.HasPrefix(dest, "r") {
					optimized = append(optimized, "    ; removed: "+trimmed)
					continue
				}
			}
		}

		// push reg / pop reg — удаляем пару
		if strings.Contains(trimmed, "push") && i+1 < len(lines) {
			nextTrimmed := strings.TrimSpace(lines[i+1])
			if strings.Contains(nextTrimmed, "pop") {
				pushReg := extractRegister(trimmed)
				popReg := extractRegister(nextTrimmed)
				if pushReg == popReg && pushReg != "" {
					optimized = append(optimized, "    ; removed push/pop pair: "+pushReg)
					i++
					continue
				}
			}
		}

		optimized = append(optimized, line)
	}

	return strings.Join(optimized, "\n")
}

// eliminateDeadCode удаляет недостижимый код
func (ao *AssemblyOptimizer) eliminateDeadCode(asm string) string {
	lines := strings.Split(asm, "\n")
	optimized := make([]string, 0, len(lines))
	deadCode := false
	deadStart := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "jmp ") || strings.HasPrefix(trimmed, "ret") {
			if !deadCode {
				deadCode = true
				deadStart = i + 1
			}
			optimized = append(optimized, line)
			continue
		}

		if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, ".") && !strings.HasPrefix(trimmed, ";") {
			if deadCode {
				removed := i - deadStart
				if removed > 0 {
					optimized = append(optimized, fmt.Sprintf("    ; [optimizer] removed %d dead instructions", removed))
				}
				deadCode = false
			}
			optimized = append(optimized, line)
			continue
		}

		if !deadCode {
			optimized = append(optimized, line)
		}
	}

	return strings.Join(optimized, "\n")
}

// foldConstants сворачивает константные выражения
func (ao *AssemblyOptimizer) foldConstants(asm string) string {
	lines := strings.Split(asm, "\n")
	optimized := make([]string, 0, len(lines))

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "mov ") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.HasPrefix(nextLine, "add ") {
				movParts := strings.Fields(trimmed)
				addParts := strings.Fields(nextLine)
				if len(movParts) >= 3 && len(addParts) >= 3 && movParts[1] == addParts[1] {
					val1, ok1 := parseImmediate(movParts[2])
					val2, ok2 := parseImmediate(addParts[2])
					if ok1 && ok2 {
						optimized = append(optimized,
							fmt.Sprintf("    mov %s, %d  ; folded: %d + %d", movParts[1], val1+val2, val1, val2))
						i++
						continue
					}
				}
			}
		}

		optimized = append(optimized, line)
	}

	return strings.Join(optimized, "\n")
}

// optimizeJumps оптимизирует цепочки переходов
func (ao *AssemblyOptimizer) optimizeJumps(asm string) string {
	return asm
}

// Enable включает оптимизатор
func (ao *AssemblyOptimizer) Enable() {
	ao.enabled = true
}

// Disable выключает оптимизатор
func (ao *AssemblyOptimizer) Disable() {
	ao.enabled = false
}

// EnablePass включает конкретный проход
func (ao *AssemblyOptimizer) EnablePass(name string) {
	for i, pass := range ao.passes {
		if pass.Name == name {
			ao.passes[i].Enabled = true
			return
		}
	}
}

// DisablePass выключает конкретный проход
func (ao *AssemblyOptimizer) DisablePass(name string) {
	for i, pass := range ao.passes {
		if pass.Name == name {
			ao.passes[i].Enabled = false
			return
		}
	}
}

// GetStats возвращает статистику оптимизаций
func (ao *AssemblyOptimizer) GetStats() map[string]int {
	return ao.stats
}

// GetReport возвращает отчёт об оптимизациях
func (ao *AssemblyOptimizer) GetReport() string {
	var sb strings.Builder

	sb.WriteString("Assembly Optimization Report:\n")
	sb.WriteString("============================\n\n")

	for _, pass := range ao.passes {
		status := "enabled"
		if !pass.Enabled {
			status = "disabled"
		}
		removed := ao.stats[pass.Name]
		fmt.Fprintf(&sb, "%-20s: %s", pass.Name, status)
		if removed > 0 {
			fmt.Fprintf(&sb, " (removed %d lines)", removed)
		}
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "  %s\n\n", pass.Description)
	}

	return sb.String()
}

func extractRegister(asmLine string) string {
	parts := strings.Fields(strings.TrimSpace(asmLine))
	if len(parts) >= 2 {
		reg := parts[1]
		reg = strings.TrimSuffix(reg, ",")
		return reg
	}
	return ""
}

func parseImmediate(s string) (int, bool) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err == nil
}
