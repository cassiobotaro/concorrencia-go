package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

// replica simula um servidor cuja latência varia a cada chamada. Se o
// contexto for cancelado antes da resposta, ela desiste, como faria uma
// requisição HTTP feita com http.NewRequestWithContext.
func replica(nome string) func(context.Context, string) (string, error) {
	return func(ctx context.Context, consulta string) (string, error) {
		select {
		case <-time.After(rand.N(100 * time.Millisecond)):
			return fmt.Sprintf("%s respondeu a %q", nome, consulta), nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// primeiro envia a mesma consulta a todas as réplicas e devolve a primeira
// resposta que chegar. Ao retornar, cancela o contexto das outras, para
// que as perdedoras parem de trabalhar em vez de responder à toa.
func primeiro(ctx context.Context, consulta string, replicas ...func(context.Context, string) (string, error)) (string, error) {
	ctx, cancelar := context.WithCancel(ctx)
	defer cancelar()

	// O buffer tem uma vaga por réplica. Uma perdedora pode terminar entre
	// a chegada da vencedora e o cancel, e sem a vaga ficaria presa no
	// envio para sempre, pois ninguém mais vai ler do canal.
	respostas := make(chan string, len(replicas))
	for _, r := range replicas {
		go func() {
			if resposta, err := r(ctx, consulta); err == nil {
				respostas <- resposta
			}
		}()
	}

	select {
	case resposta := <-respostas:
		return resposta, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	replicas := []func(context.Context, string) (string, error){
		replica("réplica 1"),
		replica("réplica 2"),
		replica("réplica 3"),
	}

	resposta, err := primeiro(context.Background(), "golang", replicas...)
	if err != nil {
		fmt.Println("erro:", err)
		return
	}
	fmt.Println(resposta)

	// Combinado com timeout: usa a resposta mais rápida, desde que chegue
	// em até 20ms. O prazo vem no contexto, e primeiro o repassa às réplicas.
	ctx, cancelar := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelar()
	resposta, err = primeiro(ctx, "csp", replicas...)
	if err != nil {
		fmt.Println("tempo esgotado: nenhuma réplica respondeu em 20ms")
		return
	}
	fmt.Println(resposta)
}
