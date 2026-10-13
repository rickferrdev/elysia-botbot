# Pendências

## Funcionalidades

- [ ] Implementar o tratamento de anexos JSON em `HandlerAttachments`, em `internal/core/services/embeds/embeds.go`. Criação e atualização aceitam anexos, mas o método ainda retorna que a funcionalidade não está implementada.
- [ ] Corrigir a consulta de incorporações em `Obtain`, em `internal/core/services/embeds/embeds.go`, para exibir também registros que tenham apenas `Embeds` e nenhum `Content`. Atualmente, esse caso responde sucesso sem exibir a incorporação.

## Testes

- [ ] Adicionar testes automatizados para o mapeamento de erros SQLite em `internal/outbound/databases/repoerrors` e para os fluxos de sucesso e erro dos serviços de embeds e autoreactions. O projeto ainda não possui arquivos `*_test.go`.
