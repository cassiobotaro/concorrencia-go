package main

import (
	"strings"
	"testing"
	"testing/synctest"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

// As chegadas saem em ordem, porque o Sleep escalonado é determinístico
// na bolha. A ordem de passagem é imprevisível, mas toda passagem vem
// depois da última chegada, e é isso que o teste confere.
func TestExemplo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		obtida := saida.Capturar(t, func() {
			esperarTodas(4)
		})

		linhas := strings.Split(strings.TrimSpace(obtida), "\n")
		if len(linhas) != 8 {
			t.Fatalf("esperava 8 linhas, obteve %d:\n%s", len(linhas), obtida)
		}
		saida.Conferir(t, strings.Join(linhas[:4], "\n"), `
gorrotina 1: chegou à barreira
gorrotina 2: chegou à barreira
gorrotina 3: chegou à barreira
gorrotina 4: chegou à barreira
`)
		saida.ConferirSemOrdem(t, strings.Join(linhas[4:], "\n"), `
gorrotina 1: passou a barreira
gorrotina 2: passou a barreira
gorrotina 3: passou a barreira
gorrotina 4: passou a barreira
`)
	})
}
