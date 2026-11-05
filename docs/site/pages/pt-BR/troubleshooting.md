# Solução de problemas

Comece com `migrate version`, `migrate capabilities --output json`, `migrate status --output json`. Só status precisa conectar.

Confira PATH na [instalação](./installation), gere a configuração no projeto ou use `-c` / `-w`. No YAML confira posição, valores e espaços em vez de tabs. Consulte a [ordem de busca](./configuration#configuration-file-search-order).

`can't find env value for XXX` pede `${XXX}` definida; `${XXX:default}` só para valores não secretos apropriados. Campos SSH `password_env` contêm nomes.

Para formato ou banco incompatível, confira capabilities, atualize versões antigas e revise os [aliases](./databases#aliases). Para módulos, confira caminhos, instale npm e use `DependencyDir` em fontes embutidas. Dependências Node ou assíncronas continuam incompatíveis.

Restaure arquivos e nomes aplicados a partir do Git e crie uma versão superior. Não use baseline para esconder estado desconhecido. Comandos silenciosos podem ter sucesso; confira applied, baseline e outOfOrder.

`to-version must be set, or use --all` exige destino ou `--all`; o destino é mantido. Scripts sem down não revertem.

Inspecione dados e histórico antes de repetir após falha: DDL implícito, operações sem transação e desconexões no COMMIT podem deixar mudanças. Baseline não repara isso. Confira credenciais SSH, known_hosts, proxy e acesso do bastion ao banco. Não há reconexão nem repetição automática; consulte logs controlados sem divulgar segredos ou dados.
