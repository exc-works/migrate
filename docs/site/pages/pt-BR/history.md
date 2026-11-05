# Histórico e estado inicial

O histórico registra versão, nome e hash; compare com o esquema real.

```sh
migrate status
migrate status --output json
```

| Estado | Significado | Ação |
| --- | --- | --- |
| `pending` | Não aplicado | Revisar e usar `up` |
| `applied` | Sucesso | Manter arquivo |
| `baseline` | Registrado sem execução | Preservar histórico adotado |
| `outOfOrder` | Pendente mais antigo | Conferir ordem |
| `hashMismatch` | Conteúdo alterado | Restaurar original |
| `filenameMismatch` | Nome alterado | Restaurar nome |

Divergências bloqueiam. Novas mudanças exigem novas versões. Hash de script não inclui [módulos](./modules#checksums).

## Adotar banco existente {#baseline}

Somente se esquema e dados já correspondem aos arquivos:

1. Compare arquivos e estado real.
2. Use `migrate create` se preciso.
3. Execute `migrate baseline`, depois `migrate status`.
4. Adicione versões superiores para mudanças futuras.

Baseline registra sem executar, não repara migração parcialmente falha.
