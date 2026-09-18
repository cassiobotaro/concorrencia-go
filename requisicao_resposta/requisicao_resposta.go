package main

import "fmt"

// requisicao carrega, além do valor, o canal pelo qual quem pediu quer
// receber a resposta. O serviço só precisa escrever nele, por isso chan<-.
type requisicao struct {
	valor    int
	resposta chan<- int
}

// servico atende uma requisição por vez e responde no canal que veio
// dentro da própria mensagem.
func servico(entrada <-chan requisicao) {
	for req := range entrada {
		req.resposta <- req.valor * 2
	}
}

func main() {
	entrada := make(chan requisicao)
	pronto := make(chan struct{})
	go func() {
		servico(entrada)
		close(pronto)
	}()

	for i := range 5 {
		resposta := make(chan int)
		entrada <- requisicao{valor: i, resposta: resposta}
		// Fica bloqueado até o serviço responder
		fmt.Println("resposta:", <-resposta)
	}

	// Sem mais requisições: o serviço termina
	close(entrada)
	<-pronto
}
