package main

import (
	"fmt"
	"time"
)

// comLatencia cria uma réplica com latência fixa, para que o teste saiba
// de antemão quem responde primeiro.
func comLatencia(nome string, latencia time.Duration) func(string) string {
	return func(consulta string) string {
		time.Sleep(latencia)
		return fmt.Sprintf("%s respondeu a %q", nome, consulta)
	}
}

func Example() {
	fmt.Println(primeiro("golang",
		comLatencia("réplica lenta", 300*time.Millisecond),
		comLatencia("réplica rápida", 10*time.Millisecond),
		comLatencia("réplica média", 150*time.Millisecond),
	))

	// Output:
	// réplica rápida respondeu a "golang"
}
