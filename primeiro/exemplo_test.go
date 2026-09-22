package main

import (
	"context"
	"fmt"
	"time"
)

// comLatencia cria uma réplica com latência fixa, para que o teste saiba
// de antemão quem responde primeiro.
func comLatencia(nome string, latencia time.Duration) func(context.Context, string) (string, error) {
	return func(ctx context.Context, consulta string) (string, error) {
		select {
		case <-time.After(latencia):
			return fmt.Sprintf("%s respondeu a %q", nome, consulta), nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func Example() {
	replicas := []func(context.Context, string) (string, error){
		comLatencia("réplica lenta", 300*time.Millisecond),
		comLatencia("réplica rápida", 10*time.Millisecond),
		comLatencia("réplica média", 150*time.Millisecond),
	}

	resposta, err := primeiro(context.Background(), "golang", replicas...)
	fmt.Println(resposta, err)

	// Com um prazo menor do que a réplica mais rápida, ninguém responde
	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelar()
	resposta, err = primeiro(ctx, "golang", replicas...)
	fmt.Printf("%q %v\n", resposta, err)

	// Output:
	// réplica rápida respondeu a "golang" <nil>
	// "" context deadline exceeded
}
