// Package saida captura o que um exemplo imprime, para conferir a saída em
// testes que rodam dentro de synctest.Test. Uma função Example faria isso
// sozinha, mas ela não recebe um *testing.T e por isso não entra na bolha.
package saida

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// Capturar executa f com os.Stdout apontando para um arquivo temporário e
// devolve o que foi escrito. Um arquivo, e não um os.Pipe: o pipe pediria
// uma gorrotina lendo em paralelo, e uma gorrotina parada em leitura de
// pipe não conta como bloqueada para a bolha, o que travaria o relógio.
func Capturar(t *testing.T, f func()) string {
	t.Helper()
	arquivo, err := os.CreateTemp(t.TempDir(), "saida")
	if err != nil {
		t.Fatal(err)
	}
	defer arquivo.Close()

	original := os.Stdout
	os.Stdout = arquivo
	defer func() { os.Stdout = original }()
	f()

	conteudo, err := os.ReadFile(arquivo.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(conteudo)
}

// Conferir compara a saída linha a linha, como o "// Output:" de um Example.
func Conferir(t *testing.T, obtida, esperada string) {
	t.Helper()
	if o, e := linhas(obtida), linhas(esperada); !slices.Equal(o, e) {
		t.Errorf("saída obtida:\n%s\nsaída esperada:\n%s", strings.Join(o, "\n"), strings.Join(e, "\n"))
	}
}

// ConferirSemOrdem compara só o conjunto de linhas, como o
// "// Unordered output:" de um Example.
func ConferirSemOrdem(t *testing.T, obtida, esperada string) {
	t.Helper()
	o, e := linhas(obtida), linhas(esperada)
	slices.Sort(o)
	slices.Sort(e)
	if !slices.Equal(o, e) {
		t.Errorf("saída obtida:\n%s\nsaída esperada:\n%s", strings.Join(o, "\n"), strings.Join(e, "\n"))
	}
}

func linhas(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}
