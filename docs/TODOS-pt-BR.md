# Pendências

## Funcionalidades

- [ ] Implementar a leitura e validação de anexos JSON em `HandlerAttachments`, em `internal/core/services/embeds/embeds.go`. O método retorna explicitamente que o tratamento não está implementado, embora os fluxos de criação e atualização aceitem anexos.
- [ ] Exibir incorporações salvas sem texto em `Obtain`, em `internal/core/services/embeds/embeds.go`. A resposta atual só edita a interação quando `decode.Content` não está vazio; incorporações contendo apenas `decode.Embeds` não são exibidas.

## Testes

- [ ] Criar testes automatizados para os serviços de embeds e autoreactions, cobrindo os fluxos de sucesso e os erros de persistência. O projeto não possui arquivos `*_test.go` no momento.
