package main

import "fmt"

// Com uma única vaga, nunca pode haver mais de uma tarefa ativa: se o
// semáforo falhasse, alguma linha mostraria "ativas: 2". Na versão com
// errgroup, a vaga única também torna a ordem previsível: a tarefa 3
// falha e as duas seguintes nem começam.
func Example() {
	executarTarefas(3, 1)

	if err := executarTarefasErrgroup(5, 1, 3); err != nil {
		fmt.Println("errgroup:", err)
	}

	// Unordered output:
	// tarefa  1 começou, ativas: 1
	// tarefa  2 começou, ativas: 1
	// tarefa  3 começou, ativas: 1
	// errgroup: tarefa  1 começou, ativas: 1
	// errgroup: tarefa  2 começou, ativas: 1
	// errgroup: tarefa  3 começou, ativas: 1
	// errgroup: tarefa  4 cancelada
	// errgroup: tarefa  5 cancelada
	// errgroup: tarefa 3 falhou
}
