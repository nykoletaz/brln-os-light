# LightningOS Light

<img width="1920" height="1080" alt="logo" src="https://github.com/user-attachments/assets/504ec23e-31f8-407a-a848-3fa4ce3ec1f9" />

[Clique aqui](https://github.com/jvxis/brln-os-light/blob/main/README.md) para ver a versão em inglês (fonte da verdade).

LightningOS Light é um instalador completo de daemon de nó Lightning, com gerenciador de nó, assistente guiado, dashboard e carteira. O manager serve UI e API via HTTPS em `0.0.0.0:8443` por padrão para acesso em LAN (defina `server.host: "127.0.0.1"` para acesso somente local) e integra com systemd, Postgres, smartctl, Tor/i2pd e LND gRPC.

<img width="1494" height="1045" alt="image" src="https://github.com/user-attachments/assets/8fb801c0-4946-48d8-8c24-c36a53d193b3" />
<img width="1491" height="903" alt="image" src="https://github.com/user-attachments/assets/cfda34d5-bccc-4b18-9970-bad494ae77b3" />
<img width="1576" height="1337" alt="image" src="https://github.com/user-attachments/assets/019cfff2-f354-4c2b-a595-2a15bb228864" />
<img width="1280" height="660" alt="image" src="https://github.com/user-attachments/assets/84489b07-8397-4195-b0d4-7e332618666d" />

## Destaques
- Apenas Mainnet (Bitcoin remoto por padrão)
- Sem Docker na stack principal
- LND gerenciado via systemd, gRPC em localhost
- A seed phrase nunca é persistida nem registrada em logs
- Assistente para credenciais RPC do Bitcoin e setup de carteira
- Suite Lightning Ops: peers/canais, Rebalance Center, Autofee, sinais HTLC e Channel Auto Heal
- Chat Keysend: 1 sat por mensagem + taxas de roteamento, indicadores de não lidas, retenção de 30 dias
- Notificações em tempo real (on-chain, Lightning, canais, forwards, rebalances)
- Notificações Telegram: backups SCB, resumos financeiros, comandos sob demanda `/scb` e `/balances`
- Relatórios diários de roteamento (timer + backfill + API live)
- App Store: LNDg, Peerswap (psweb), Elements, Bitcoin Core
- Gestão de Bitcoin Local (status + config) e visualizador de logs

## Estrutura do repositório
- `cmd/lightningos-manager`: backend Go (API + UI estática)
- `ui`: UI React + Tailwind
- `templates`: units systemd e templates de configuração
- `install.sh`: instalador idempotente (wrapper em `scripts/install.sh`)
- `configs/config.yaml`: configuração local de desenvolvimento

## Instalação (Ubuntu Server)
O instalador provisiona tudo que é necessário em um Ubuntu limpo:
- Postgres, smartmontools, curl, jq, ca-certificates, openssl, build tools
- Tor (ControlPort habilitado) + i2pd habilitado por padrão
- Go 1.22.x e Node.js 20.x (se ausentes ou antigos)
- Binários do LND (padrão `v0.20.0-beta`)
- Binário do LightningOS Manager (compilado localmente)
- Build da UI (compilada localmente)
- Serviços systemd e templates de configuração
- Certificado TLS autoassinado

Uso:
```bash
git clone https://github.com/jvxis/brln-os-light
cd brln-os-light/lightningos-light
sudo ./install.sh
```

Se você já clonou e está em `brln-os-light`, use:
```bash
cd lightningos-light
sudo ./install.sh
```

### Instalação via curl (bootstrap)
Isso baixa o repo (ou executa `git pull` se já existir) e depois roda `lightningos-light/install.sh`.
```bash
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo ACCEPT_MIT_LICENSE=1 bash
```

Overrides opcionais:
```bash
# Usar outro caminho de clone
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo BRLN_DIR=/opt/brln-os-light bash

# Fixar branch/tag
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo BRLN_BRANCH=main bash

# Usar outra URL de repositório
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo REPO_URL=https://github.com/jvxis/brln-os-light bash
```

Nota de UFW (App Store/LNDg):
Se o LNDg não alcançar o LND gRPC e o UFW estiver ativo, tráfego Docker-to-host pode ser bloqueado.
Rode os checks abaixo e permita a interface bridge usada pela rede do LNDg:
```bash
sudo docker exec -it lndg-lndg-1 getent hosts host.docker.internal
sudo docker exec -it lndg-lndg-1 bash -lc 'timeout 3 bash -lc "</dev/tcp/host.docker.internal/10009" && echo OK || echo FAIL'
sudo docker network inspect lndg_default --format '{{.Id}}'
# nome da bridge = br-<primeiros 12 caracteres do id>
sudo ufw allow in on br-<id> to any port 10009 proto tcp
```
Se ainda falhar, tente:
```bash
sudo iptables -I INPUT -i br-<id> -p tcp --dport 10009 -j ACCEPT
```

**Atenção (nós existentes):** Se você já tem um nó Lightning com LND/Bitcoin rodando, não use `install.sh`.
Siga o guia de Nó Existente:
- PT-BR: `docs/13_EXISTING_NODE_GUIDE_PT_BR.md`
- EN: `docs/14_EXISTING_NODE_GUIDE_EN.md`

Acesse a UI de outra máquina na mesma LAN:
`https://<IP_LAN_DO_SERVIDOR>:8443`

Notas:
- Você pode sobrescrever a URL do LND com `LND_URL=...` ou a versão com `LND_VERSION=...`.
- O instalador gera uma role no Postgres e atualiza `LND_PG_DSN` em `/etc/lightningos/secrets.env`.
- O rótulo de versão da UI vem de `ui/public/version.txt`.
- PostgreSQL usa o repositório PGDG por padrão. Defina `POSTGRES_VERSION=18` (ou outra major) para sobrescrever.
- Tor usa o repositório Tor Project quando disponível. Se o codinome Ubuntu não for suportado, faz fallback para `jammy`.

## Permissões do instalador (o que `install.sh` aplica)
- Usuários:
  - `lnd` (usuário de sistema, dono de `/data/lnd`)
  - `lightningos` (usuário de sistema, executa o manager service)
- Grupos:
  - `lightningos` nos grupos `lnd` e `systemd-journal`
  - `lnd` no grupo `debian-tor`
- Caminhos-chave:
  - `/etc/lightningos` e `/etc/lightningos/tls`: `root:lightningos`, `chmod 750`
  - `/etc/lightningos/secrets.env`: `root:lightningos`, `chmod 660`
  - `/data/lnd`: `lnd:lnd`, `chmod 750`
  - `/data/lnd/data/chain/bitcoin/mainnet`: `lnd:lnd`, `chmod 750`
  - `/data/lnd/data/chain/bitcoin/mainnet/admin.macaroon`: `lnd:lnd`, `chmod 640`

## Caminhos de configuração
- `/etc/lightningos/config.yaml`
- `/etc/lightningos/secrets.env` (chmod 660)
- `/data/lnd/lnd.conf`
- `/data/lnd` (diretório de dados LND)

## Notificações e backups
LightningOS Light inclui um sistema de notificações em tempo real que rastreia:
- Transações on-chain (recebidas/enviadas)
- Invoices Lightning (liquidadas) e pagamentos (enviados)
- Eventos de canal (abertura, fechamento, pendente)
- Forwards e rebalances

Notificações são armazenadas em um Postgres dedicado (veja `NOTIFICATIONS_PG_DSN` em `/etc/lightningos/secrets.env`).

## Chat (Keysend)
O chat Keysend está disponível na UI e mira apenas peers online.
- Cada mensagem envia 1 sat + taxas de roteamento.
- Mensagens são armazenadas localmente em `/var/lib/lightningos/chat/messages.jsonl` e retidas por 30 dias.
- Peers com não lidas ficam destacados até o chat ser aberto.

Notificações Telegram:
- Configure na UI: Notifications -> Telegram.
- A UI inclui um card de regras gerais com defaults operacionais.
- Backup SCB em abertura/fechamento de canal (toggle).
- Resumo financeiro agendado (intervalos de 1 a 12 horas).
- Comandos sob demanda: `/scb` (backup) e `/balances` (resumo).
- `/scb` e `/balances` são auto-registrados no menu do bot do Telegram.
- Mensagens de backup SCB incluem contexto do alias do peer na legenda.
- Token do bot vem do @BotFather e chat id vem do @userinfobot.
- Somente chat direto; deixar ambos os campos vazios desativa Telegram.

Chaves de ambiente:
- `NOTIFICATIONS_TG_BOT_TOKEN`
- `NOTIFICATIONS_TG_CHAT_ID`

## Relatórios
Relatórios diários de roteamento são calculados à meia-noite no horário local e armazenados no Postgres (mesmo DB/usuário de notificações).

Agenda:
- `lightningos-reports.timer` executa `lightningos-reports.service` às `00:00` local.
- Execução manual: `/opt/lightningos/manager/lightningos-manager reports-run --date YYYY-MM-DD` (padrão: ontem).
- Backfill: `/opt/lightningos/manager/lightningos-manager reports-backfill --from YYYY-MM-DD --to YYYY-MM-DD` (máximo padrão de 730 dias; use `--max-days N` para sobrescrever).
- Pin opcional de timezone: defina `REPORTS_TIMEZONE=America/Sao_Paulo` em `/etc/lightningos/secrets.env` para forçar relatórios diários, backfill e live no mesmo timezone IANA.

Tabela armazenada: `reports_daily`
- `report_date` (DATE, dia local)
- `forward_fee_revenue_sats`
- `forward_fee_revenue_msat`
- `rebalance_fee_cost_sats`
- `rebalance_fee_cost_msat`
- `net_routing_profit_sats`
- `net_routing_profit_msat`
- `forward_count`
- `rebalance_count`
- `routed_volume_sats`
- `routed_volume_msat`
- `onchain_balance_sats`
- `lightning_balance_sats`
- `total_balance_sats`
- `created_at`, `updated_at`

Endpoints de API:
- `GET /api/reports/range?range=d-1|month|3m|6m|12m|all` (month = últimos 30 dias)
- `GET /api/reports/custom?from=YYYY-MM-DD&to=YYYY-MM-DD` (máx. 730 dias)
- `GET /api/reports/summary?range=...`
- `GET /api/reports/live` (hoje 00:00 local -> agora, cache ~60s)

## Lightning Ops (mapa de funcionalidades)
- Gestão de canais: controles de peer/canal, atualizações de policy e refinamentos de card/saldo de canal.
- Rebalance Center: rebalances manuais + automáticos com targeting por score, watchdogs, pre-probing, guardrails de ROI e auto-restart opcional no modo manual.
- Autofee: automação de taxas por canal com âncoras de custo, seed Amboss, integração de sinais HTLC, calibração por tamanho/liquidez do nó, scheduler/manual run e histórico detalhado.
- HTLC Manager: telemetria HTLC com histerese usada pelo Autofee e por decisões de liquidez.
- Channel Auto Heal + Tor peers checker: guardrails operacionais para confiabilidade de peer/canal.
- Health checks: opção de follow-bitcoin para fluxos de saúde de LND/nó.

## Rebalance Center
Rebalance Center é um otimizador de liquidez de entrada (local/outbound) para LND. Ele pode rodar rebalances manuais por canal ou varreduras totalmente automáticas que enfileiram rebalances com base em ROI e restrições de orçamento. Um rebalance só avança quando **outgoing fee > peer fee**, para que você nunca pague mais do que a cobrança do peer sem spread positivo. Custos são rastreados por notificações (fee msat) e agregados em custo live + gasto diário auto/manual.

Comportamento principal:
- Rebalances manuais ignoram orçamento diário e podem ser iniciados por canal.
- Rebalances automáticos respeitam orçamento diário e só miram canais marcados explicitamente como `Auto`.
- Canais de origem são selecionados entre os com liquidez local suficiente e não excluídos; um canal preenchido por rebalance fica **protegido** e não pode ser usado como origem até regras de payback liberarem.
- Alvos são escolhidos quando o déficit de liquidez outbound passa do deadband e o spread de taxa é positivo; estimativa de ROI usa receita de roteamento dos últimos 7 dias vs custo estimado de rebalance.
- Alvos automáticos são ranqueados por **economic score** = (ganho esperado - custo estimado), priorizando canais de maior margem.
- Um **profit guardrail** impede enqueue automático quando ganho esperado é menor que custo estimado (quando ambos são conhecidos). Se ROI for indeterminado (cost = 0 com spread positivo), auto continua permitido.
- Seleção de origem é ponderada por histórico do par: pares recentes bem-sucedidos com taxas menores são priorizados, e falhas recentes são despriorizadas.
- A visão geral mostra **Last scan** em horário local e status da varredura (ex.: sem origens, sem candidatos, orçamento esgotado), além de telemetria econômica (top score, skips por profit guardrail) e detalhes opcionais de skip.
- Rebalances manuais podem opcionalmente usar **auto-restart** (toggle por canal) com cooldown de 60s até o alvo ser alcançado.
- **Pre-probing** de rota roda antes do envio, buscando o maior valor viável na rota.

Channel Workbench:
- Define percentual-alvo de outbound por canal.
- Toggle `Auto` para permitir que auto mode rebalanceie o canal.
- Toggle no ícone de restart para auto-restart de rebalances manuais do canal.
- Toggle `Exclude source` para bloquear um canal como origem.
- Ordenação: **Economic** (baseada em score) ou **Emptiest** (menor % local primeiro).

Codificação por cor (linhas de canal):
- Fundo verde = origem elegível (pode financiar rebalances).
- Fundo vermelho = alvo elegível (auto habilitado e precisando de outbound).
- Fundo âmbar = alvo potencial (precisa de outbound, mas auto desabilitado).

Parâmetros de configuração:
- Configurações somente auto: `Enable auto rebalance`, `Scan interval (sec)`, `Daily budget (% of revenue)`.
- `Enable auto rebalance`: liga/desliga varredura automática.
- `Scan interval (sec)`: frequência da varredura automática.
- `Daily budget (% of revenue)`: percentual da receita de roteamento das últimas 24h alocado para auto-rebalances.
- `Deadband (%)`: déficit mínimo de outbound antes de um canal virar alvo.
- `Minimum local for source (%)`: liquidez local mínima para um canal ser origem.
- `Economic ratio`: fração da taxa outbound do canal alvo (base+ppm) usada como limite máximo de taxa.
- `Econ ratio max (ppm)`: teto opcional para o limite de taxa ao usar economic ratio (0 = sem teto).
- `Fee limit (ppm)`: sobrescreve economic ratio com limite fixo máximo de taxa ppm (0 = desativado).
- `Subtract source fees`: reduz orçamento de taxa com estimativa de source fees (mais conservador).
- `ROI minimum`: ROI mínimo estimado (receita 7d / custo estimado) para enfileirar jobs auto.
- `Max concurrent`: máximo de rebalances simultâneos.
- `Minimum (sats)`: menor valor de rebalance para tentativas padrão (probing pode ir abaixo para capturar rota válida).
- `Maximum (sats)`: limite superior do tamanho de rebalance (0 = ilimitado).
- `Fee ladder steps`: número de fee caps tentados do menor para o maior antes de desistir.
- `Amount probe steps`: número de sondas de valor (do maior para o menor) quando ocorre falha temporária no último salto.
- `Fail tolerance (ppm)`: probing para quando delta entre valores ficar abaixo desse limite.
- `Adaptive amount probing`: limita a próxima tentativa com base no último valor bem-sucedido.
- `Attempt timeout (sec)`: tempo máximo por tentativa antes de seguir para próxima taxa/valor.
- `Rebalance timeout (sec)`: tempo máximo por job de rebalance (auto ou manual).
- `Mission control half-life (sec)`: tempo de decaimento de falhas no mission control (menor = esquece mais rápido, 0 = padrão do LND).
- `Payback policy`: três modos podem ser habilitados juntos.
- `Release by payback`: libera liquidez protegida quando receita de roteamento paga o custo do rebalance.
- `Release by time`: libera após `Unlock days` desde o último rebalance.
- `Critical mode`: libera uma fração quando origens ficam escassas por várias varreduras.
- `Unlock days`: número de dias para desbloqueio por tempo.
- `Critical release (%)`: percentual de liquidez protegida liberada por ciclo crítico.
- `Critical cycles`: varreduras consecutivas com poucas origens antes de acionar liberação crítica.
- `Critical min sources`: mínimo de canais origem elegíveis para evitar modo crítico.
- `Critical min available sats`: liquidez total mínima de origem para evitar modo crítico.

## Lightning Ops: Autofee
Autofee ajusta automaticamente **outbound fees** para maximizar **lucro primeiro** e **movimento depois**. Ele usa histórico local de roteamento e rebalance (notificações no Postgres) + métricas opcionais Amboss para seed de preço, e depois aplica guardrails, cooldowns e caps para manter updates seguros e explicáveis.

Parâmetros da UI:
- `Enable autofee`: liga/desliga global.
- `Profile`: Conservative / Moderate / Aggressive (define limites internos).
- `Lookback window (days)`: 5 a 21 dias para estatísticas.
- `Run interval (hours)`: mínimo de 1 hora.
- `Cooldown up / down (hours)`: tempo mínimo entre aumentos/reduções de fee.
- `Min fee (ppm)` e `Max fee (ppm)`: clamps rígidos (mínimo pode ser `0`).
- `Rebalance cost mode`: `Per-channel`, `Global` ou `Blend` (controla a âncora de custo usada em floors/margens).
- `Amboss fee reference`: seed opcional; requer token de API.
- `Inbound passive rebalance`: usa desconto inbound para canais sink.
- `Discovery mode`: redução mais rápida para canais ociosos/alto outbound.
- `Explorer mode`: ciclos temporários de exploração; pode pular cooldown em movimentos de queda.
- `Revenue floor`: mantém floor mínimo para canais de alta performance.
- `Circuit breaker`: reduz steps se demanda cair após aumentos recentes.
- `Extreme drain`: acelera aumento de fee quando canal está cronicamente drenado.
- `Super source` + base fee: aumenta base fee quando canal é classificado como super source.
- `HTLC signal integration`: habilita feedback de falhas HTLC do HTLC Manager.
- `HTLC mode`: `observe_only` (só telemetria/tags), `policy_only` (só efeitos de policy), `full` (policy + efeitos de liquidez, padrão).

Comportamento de sinais HTLC:
- Janela de sinal alinhada à cadência do Autofee: `max(run_interval, 60m)`.
- Limites mínimos de amostra/falha são escalados pela janela HTLC ativa e calibrados por tamanho do nó + classe de liquidez.
- Linha de resumo inclui: `htlc_liq_hot`, `htlc_policy_hot`, `htlc_low_sample`, `htlc_window`.
- Linha por canal inclui contadores quando presentes: `htlc<window>m a=<attempts> p=<policy_fails> l=<liquidity_fails>`.

Calibração automática:
- Cada execução calcula classificação do nó e status de liquidez para autoescalar limites.
- Classes de tamanho do nó (capacidade total + quantidade de canais):
  - `small`: < 50M sats ou < 20 canais
  - `medium`: < 200M sats ou < 60 canais
  - `large`: < 1.5B sats ou < 150 canais
  - `extra large`: acima disso
- Classes de liquidez (local ratio):
  - `drained`: local ratio < 25%
  - `balanced`: 25% a 75%
  - `full`: local ratio > 75%
- A linha de calibração em Autofee Results mostra essas classes + limites `revfloor` calibrados.

Linhas de Autofee Results:
- Header: tipo de execução e timestamp.
- Summary: contagens de up/down/flat e motivos de skip.
- Seed: uso de Amboss e fallback.
- Calibration: tamanho do nó, liquidez e limites calibrados.
- Linhas por canal: decisão, alvo, floors, margens e tags.
- Filtros de resultado: exibir últimas N execuções e filtrar opcionalmente por intervalo de tempo local.

Glossário de tags (Autofee Results):
- `sink`, `source`, `router`, `unknown`: rótulos de classe.
- `discovery`: canal em discovery mode.
- `discovery-hard`: harddrop de discovery acionado (sem baseline + ocioso).
- `explorer`: explorer mode ativo.
- `cooldown-skip`: cooldown ignorado em movimento de queda por causa do explorer.
- `surge+X%`: surge bump aplicado.
- `top-rev`: bump de top-revenue-share aplicado.
- `neg-margin`: proteção de margem negativa aplicada.
- `negm+X%`: uplift adicional de margem negativa.
- `revfloor`: revenue floor aplicado.
- `outrate-floor`: outrate floor aplicado.
- `peg`: peg para outrate observada.
- `peg-grace`: peg aplicado dentro da grace window.
- `peg-demand`: peg aplicado por demanda forte vs seed.
- `circuit-breaker`: circuit breaker reduziu o step.
- `extreme-drain`: caminho de boost de step por extreme drain usado.
- `extreme-drain-turbo`: uplift turbo extra de extreme drain.
- `cooldown`: cooldown bloqueou update.
- `cooldown-profit`: cooldown segurou queda lucrativa.
- `hold-small`: mudança abaixo do delta/percentual mínimo.
- `same-ppm`: target igual ao ppm atual.
- `no-down-low`: queda bloqueada enquanto canal está fortemente drenado.
- `no-down-neg-margin`: queda bloqueada enquanto margem está negativa.
- `sink-floor`: margem de floor extra aplicada a canais sink.
- `trend-up`, `trend-down`, `trend-flat`: dica direcional do próximo movimento.
- `stepcap`: step cap aplicado.
- `stepcap-lock`: lock de step cap aplicado.
- `floor-lock`: floor lock aplicado.
- `global-neg-lock`: lock global de margem negativa aplicado.
- `lock-skip-no-chan-rebal`: lock ignorado por falta de histórico de rebalance por canal.
- `lock-skip-sink-profit`: lock ignorado no caminho de lucratividade de sink.
- `profit-protect-lock`: lock de proteção de lucro aplicado.
- `profit-protect-relax`: proteção de lucro relaxada.
- `super-source`: canal classificado como super source.
- `super-source-like`: classificação super-source do tipo router-like.
- `inb-<n>`: desconto inbound aplicado.
- `htlc-policy-hot`: sinal HTLC de falhas de policy alto.
- `htlc-liquidity-hot`: sinal HTLC de falhas de liquidez alto.
- `htlc-sample-low`: amostra HTLC baixa demais para classificação hot.
- `htlc-neutral-lock`: ambos sinais hot HTLC presentes e caminho de neutral lock usado.
- `htlc-liq+X%`: bump por liquidez-hot aplicado.
- `htlc-policy+X%`: bump por policy-hot aplicado.
- `htlc-liq-nodown`: queda bloqueada por sinal liquidez-hot.
- `htlc-policy-nodown`: queda bloqueada por sinal policy-hot.
- `htlc-neutral-nodown`: queda bloqueada por lock combinado HTLC.
- `htlc-step-boost`: boost de step-cap por sinal HTLC hot.
- `seed:amboss`: seed Amboss usada.
- `seed:amboss-missing`: token Amboss ausente.
- `seed:amboss-empty`: Amboss retornou sem dados.
- `seed:amboss-error`: erro ao buscar Amboss.
- `seed:med`: mediana Amboss misturada.
- `seed:vol-<n>%`: penalidade de volatilidade aplicada.
- `seed:ratio<factor>`: ajuste por razão out/in aplicado.
- `seed:outrate`: seed de outrate recente.
- `seed:mem`: seed de memória.
- `seed:default`: fallback de seed padrão.
- `seed:guard`: guard de salto de seed aplicado.
- `seed:p95cap`: seed limitada ao p95 da Amboss.
- `seed:absmax`: cap absoluto de seed aplicado.

## Terminal web (opcional)
LightningOS Light pode expor um terminal web protegido usando GoTTY.

O instalador habilita automaticamente o terminal e gera credencial quando ausente.
Você pode revisar/sobrescrever em `/etc/lightningos/secrets.env`:
- `TERMINAL_ENABLED=1`
- `TERMINAL_CREDENTIAL=user:pass`
- `TERMINAL_ALLOW_WRITE=0` (defina `1` para permitir input)
- `TERMINAL_PORT=7681` (opcional)
- `TERMINAL_WS_ORIGIN=^https://.*:8443$` (opcional, padrão permite todas as origens)

Inicie (ou reinicie) o serviço:
```bash
sudo systemctl enable --now lightningos-terminal
```
A página Terminal mostra a senha atual e botão de cópia.

## Notas de segurança
- A seed phrase nunca é armazenada. Ela é mostrada uma vez no assistente.
- Credenciais RPC são armazenadas apenas em `/etc/lightningos/secrets.env` (root:lightningos, `chmod 660`).
- API/UI bindam em `0.0.0.0` por padrão para acesso LAN. Para localhost-only, defina `server.host: "127.0.0.1"` em `/etc/lightningos/config.yaml`.

## Troubleshooting
Se `https://<IP_LAN_DO_SERVIDOR>:8443` não estiver acessível:
```bash
systemctl status lightningos-manager --no-pager
journalctl -u lightningos-manager -n 200 --no-pager
ss -ltn | grep :8443
```

### App Store (LNDg, Peerswap, Elements, Bitcoin Core)
- LNDg roda em Docker e escuta em `http://<IP_LAN_DO_SERVIDOR>:8889`.
- Peerswap instala `peerswapd` + `psweb` (UI em `http://<IP_LAN_DO_SERVIDOR>:1984`) e requer Elements.
- Elements roda como serviço nativo (Liquid Elements node, RPC em `127.0.0.1:7041`).
- Bitcoin Core roda via Docker com dados em `/data/bitcoin`.

Notas LNDg:
- A página de logs do LNDg lê `/var/log/lndg-controller.log` dentro do container. Se estiver vazio, verifique `docker logs lndg-lndg-1`.
- Se aparecer `Is a directory: /var/log/lndg-controller.log`, remova `/var/lib/lightningos/apps-data/lndg/data/lndg-controller.log` no host e reinicie o LNDg.
- Se LND estiver usando Postgres, o LNDg pode logar ausência de `channel.db`. Isso é esperado e inofensivo.

## Arquitetura da App Store
- Cada app implementa um handler em `internal/server/apps_<app>.go`.
- Apps são registrados em `internal/server/apps_registry.go`.
- Arquivos de app ficam em `/var/lib/lightningos/apps/<app>` e dados persistentes em `/var/lib/lightningos/apps-data/<app>`.
- Docker é instalado sob demanda por apps que precisam dele (a instalação core continua sem Docker).
- Checks de sanidade de registry garantem IDs e portas únicos.

### Adicionando um novo app
1) Crie `internal/server/apps_<app>.go` e implemente a interface `appHandler`.
2) Registre o app em `internal/server/apps_registry.go`.
3) Adicione um card em `ui/src/pages/AppStore.tsx` e um ícone em `ui/src/assets/apps/`.

### Checks da App Store
Rode os testes de sanidade do registry:
```bash
go test ./internal/server -run TestValidateAppRegistry
```

## Changelog
Notas por versão são mantidas no GitHub Releases:
- https://github.com/jvxis/brln-os-light/releases

## Desenvolvimento
Veja `DEVELOPMENT.md` para setup local e instruções de build.

## Systemd
Templates estão em `templates/systemd/`.

## Rebuild apenas (manager/UI)
Use quando quiser apenas recompilar sem rodar o instalador completo.

Rebuild do manager:
```bash
sudo /usr/local/go/bin/go build -o dist/lightningos-manager ./cmd/lightningos-manager
sudo install -m 0755 dist/lightningos-manager /opt/lightningos/manager/lightningos-manager
sudo systemctl restart lightningos-manager
```

Rebuild da UI:
```bash
cd ui && sudo npm install && sudo npm run build
cd ..
sudo rm -rf /opt/lightningos/ui/*
sudo cp -a ui/dist/. /opt/lightningos/ui/
```
