package main

// Com uma única vaga, nunca pode haver mais de uma tarefa ativa: se o
// semáforo falhasse, alguma linha mostraria "ativas: 2".
func Example() {
	executarTarefas(3, 1)

	// Unordered output:
	// tarefa  1 começou, ativas: 1
	// tarefa  2 começou, ativas: 1
	// tarefa  3 começou, ativas: 1
}
