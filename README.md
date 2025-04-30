# Comparação JSON vs FlatBuffers

Este projeto compara o desempenho e o tamanho de dados serializados em JSON e FlatBuffers. Convite feito pelo professor Wesley (FullCycle) durante aula bônus sobre "Comunicação e Resiliência" onde estudamos também sobre comunicações síncronas e assíncronas.

## Objetivo
- Desempenho: Medir o tempo de escrita e leitura dos dados serializados.
- Tamanho: Comparar o tamanho dos arquivos gerados.

## Vantagens do FlatBuffers
- Desempenho superior: Leitura e escrita mais rápidas, especialmente em dados grandes.
- Tamanho reduzido: Arquivos menores em comparação com JSON.
- Zero-copy: Leitura eficiente diretamente na memória.

## Estrutura do Projeto
1.	Gerar Dados: 1 milhão de usuários com dados fictícios (nome, idade, e-mail, etc.).
2.	Benchmarking: Medição de tempo de serialização e deserialização.

## Como Rodar
1.	Clone o repositório:
```bash
git clone https://github.com/andersonbosa/demo-flatbuffers-vs-json.git
cd demo-flatbuffers-vs-json
```

2.	Instale dependências e execute via run mesmo:
```bash
go mod tidy
go run main.go
```

3.	Resultados:
- Arquivos gerados: output.json e output.bin.
- Gráficos gerados: benchmark_tempo_chart.html e benchmark_tamanho_chart.html.
```
     JSON        -> 321.24 MB | Write: 439.968458ms | Read: 2.15684125s
     FlatBuffers -> 289.83 MB | Write: 318.558ms | Read: 37.530833ms
```

## Conclusão
- FlatBuffers oferece melhor desempenho e tamanho reduzido em comparação com JSON, especialmente em grandes volumes de dados.
