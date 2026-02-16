# LightningOS Light

<img width="1920" height="1080" alt="logo" src="https://github.com/user-attachments/assets/504ec23e-31f8-407a-a848-3fa4ce3ec1f9" />

O LightningOS Light é um instalador completo do Daemon de Nó Lightning, um gerenciador de nó Lightning com um assistente guiado, painel de controle (dashboard) e carteira. O gerenciador fornece a interface de usuário (UI) e a API via HTTPS em `0.0.0.0:8443` por padrão para acesso em rede local (LAN) (defina `server.host: "127.0.0.1"` para acesso apenas local) e integra-se com systemd, Postgres, smartctl, Tor/i2pd e LND gRPC.
<img width="1494" height="1045" alt="image" src="[https://github.com/user-attachments/assets/8fb801c0-4946-48d8-8c24-c36a53d193b3](https://github.com/user-attachments/assets/8fb801c0-4946-48d8-8c24-c36a53d193b3)" />
<img width="1491" height="903" alt="image" src="[https://github.com/user-attachments/assets/cfda34d5-bccc-4b18-9970-bad494ae77b3](https://github.com/user-attachments/assets/cfda34d5-bccc-4b18-9970-bad494ae77b3)" />
<img width="1576" height="1337" alt="image" src="[https://github.com/user-attachments/assets/019cfff2-f354-4c2b-a595-2a15bb228864](https://github.com/user-attachments/assets/019cfff2-f354-4c2b-a595-2a15bb228864)" />
<img width="1280" height="660" alt="image" src="[https://github.com/user-attachments/assets/84489b07-8397-4195-b0d4-7e332618666d](https://github.com/user-attachments/assets/84489b07-8397-4195-b0d4-7e332618666d)" />

## Destaques

* Apenas Mainnet (Bitcoin remoto por padrão)
* Sem Docker na stack principal
* LND gerenciado via systemd, gRPC em localhost
* A frase semente (seed phrase) nunca é persistida ou registrada em logs
* Assistente para configuração de carteira e credenciais RPC do Bitcoin
* Operações Lightning (Lightning Ops): peers, canais e atualizações de taxas
* Chat Keysend: 1 sat por mensagem + taxas de roteamento, indicadores de não lida, retenção de 30 dias
* Notificações em tempo real (on-chain, Lightning, canais, roteamentos e rebalanceamentos)
* Notificações no Telegram: backups SCB, resumos financeiros, comandos `/scb` e `/balances` sob demanda
* App Store: LNDg, Peerswap (psweb), Elements, Bitcoin Core
* Gerenciamento do Bitcoin Local (status + configuração) e visualizador de logs

## Notas de lançamento

### 0.2.3 Beta

* Refatoração das notificações do Telegram: card de regras gerais, botão de ativação de backup SCB, resumos financeiros programados e comandos `/scb` + `/balances` sob demanda (registrados automaticamente no menu do bot).
* Backups SCB incluem o contexto do alias do peer na legenda do Telegram.
* Melhorias na UI de Operações Lightning: refinamentos nos cards de canais e barra de saldo de canais.
* Etapas de insights de HTLC no Autofee e calibração por tamanho/liquidez do nó, além de agendador/execução manual e correções de teto de taxa/semente.
* Correções e melhorias nas pontuações do Rebalance Center.
* Correções na atualização de configuração do LND e correção na instalação do PostgreSQL.
* Atualizações de localização (refinamentos na UI em português).

### 0.2.2 Beta

* Novo Gerenciador HTLC com histerese e execuções baseadas em minutos, além de múltiplas melhorias.
* Aprimoramentos no Rebalance Center: fluxos de reinício manual, roteamento pré-sonda (pre-probe), watchdogs, melhorias na lógica de ROI e visualização de detalhes.
* Autofee Integrado: espelha o brln-autofee, ativação por canal, suporte a taxa zero, fallback de custo de rebalanceamento, pesquisa de resultados, persistência da última execução, tendência de tags, limite de passo (step-cap) por canal, modo de relaxamento (relax mode) e filtragem de simulação (dry-run).
* Atualização do Channel Auto Heal e verificador de peers Tor.
* Correção na limpeza de atividades da carteira (remoção de saldos do histórico da carteira).
* Correções de atualização do LND e opção de acompanhar o bitcoin no health check.

## Layout do repositório

* `cmd/lightningos-manager`: Backend em Go (API + UI estática)
* `ui`: UI em React + Tailwind
* `templates`: Units do systemd e templates de configuração
* `install.sh`: Instalador idempotente (wrapper em `scripts/install.sh`)
* `configs/config.yaml`: Configuração local de desenvolvimento

## Instalação (Ubuntu Server)

O instalador provisiona tudo o que é necessário em uma máquina Ubuntu limpa:

* Postgres, smartmontools, curl, jq, ca-certificates, openssl, ferramentas de build (build tools)
* Tor (ControlPort ativado) + i2pd ativado por padrão
* Go 1.22.x e Node.js 20.x (se estiverem faltando ou muito antigos)
* Binários do LND (padrão `v0.20.0-beta`)
* Binário do LightningOS Manager (compilado localmente)
* Build da UI (compilado localmente)
* Serviços do systemd e templates de configuração
* Certificado TLS autoassinado

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

Isso baixa o repositório (ou roda `git pull` se ele já existir) e, em seguida, executa `lightningos-light/install.sh`.

```bash
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo ACCEPT_MIT_LICENSE=1 bash

```

Substituições opcionais:

```bash
# Usar um local de clone diferente
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo BRLN_DIR=/opt/brln-os-light bash

# Fixar uma branch/tag específica
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo BRLN_BRANCH=main bash

# Usar uma URL de repositório diferente
curl -fsSL https://raw.githubusercontent.com/jvxis/brln-os-light/main/lo_bootstrap.sh | sudo REPO_URL=https://github.com/jvxis/brln-os-light bash

```

Nota sobre o UFW (App Store/LNDg):
Se o LNDg falhar ao alcançar o LND gRPC e o UFW estiver habilitado, o tráfego do Docker para o host pode estar bloqueado.
Execute as seguintes verificações e permita a interface de ponte (bridge) usada pela rede do LNDg:

```bash
sudo docker exec -it lndg-lndg-1 getent hosts host.docker.internal
sudo docker exec -it lndg-lndg-1 bash -lc 'timeout 3 bash -lc "</dev/tcp/host.docker.internal/10009" && echo OK || echo FAIL'
sudo docker network inspect lndg_default --format '{{.Id}}'
# nome da ponte = br-<primeiros 12 caracteres do id>
sudo ufw allow in on br-<id> to any port 10009 proto tcp

```

Se ainda falhar, tente:

```bash
sudo iptables -I INPUT -i br-<id> -p tcp --dport 10009 -j ACCEPT

```

**Atenção (nós existentes):** Se você já tem um nó Lightning com LND/Bitcoin em execução, não use o `install.sh`.

Siga o Guia de Nó Existente em vez disso:

* PT-BR: `docs/13_EXISTING_NODE_GUIDE_PT_BR.md`
* EN: `docs/14_EXISTING_NODE_GUIDE_EN.md`

Acesse a UI de outra máquina na mesma LAN:
`https://<IP_DA_LAN_DO_SERVIDOR>:8443`

Notas:

* Você pode substituir a URL do LND usando `LND_URL=...` ou a versão usando `LND_VERSION=...`.
* O instalador irá gerar uma role (função) no Postgres e atualizará o `LND_PG_DSN` em `/etc/lightningos/secrets.env`.
* A etiqueta de versão da UI vem de `ui/public/version.txt`.
* O PostgreSQL usa o repositório PGDG por padrão. Defina `POSTGRES_VERSION=18` (ou outra versão principal) para substituir.
* O Tor usa o repositório do Tor Project quando disponível. Se o codinome do seu Ubuntu não for suportado, o fallback será para `jammy`.

## Permissões do instalador (o que o `install.sh` aplica)

* Usuários:
* `lnd` (usuário de sistema, dono de `/data/lnd`)
* `lightningos` (usuário de sistema, roda o serviço do manager)


* Associações de grupos:
* `lightningos` em `lnd` e `systemd-journal`
* `lnd` em `debian-tor`


* Caminhos chave:
* `/etc/lightningos` e `/etc/lightningos/tls`: `root:lightningos`, `chmod 750`
* `/etc/lightningos/secrets.env`: `root:lightningos`, `chmod 660`
* `/data/lnd`: `lnd:lnd`, `chmod 750`
* `/data/lnd/data/chain/bitcoin/mainnet`: `lnd:lnd`, `chmod 750`
* `/data/lnd/data/chain/bitcoin/mainnet/admin.macaroon`: `lnd:lnd`, `chmod 640`



## Caminhos de configuração

* `/etc/lightningos/config.yaml`
* `/etc/lightningos/secrets.env` (chmod 660)
* `/data/lnd/lnd.conf`
* `/data/lnd` (Diretório de dados do LND)

## Notificações e backups

O LightningOS Light inclui um sistema de notificações em tempo real que rastreia:

* Transações on-chain (recebidas/enviadas)
* Invoices Lightning (liquidadas) e pagamentos (enviados)
* Eventos de canais (abertura, fechamento, pendentes)
* Roteamentos (forwards) e rebalanceamentos

As notificações são armazenadas num banco de dados Postgres dedicado (veja `NOTIFICATIONS_PG_DSN` em `/etc/lightningos/secrets.env`).

## Chat (Keysend)

O chat Keysend está disponível na UI e destina-se apenas a peers online.

* Toda mensagem envia 1 sat + taxas de roteamento.
* As mensagens são armazenadas localmente em `/var/lib/lightningos/chat/messages.jsonl` e retidas por 30 dias.
* Peers não lidos ficam destacados até que seu chat seja aberto.

Notificações no Telegram:

* Configure na UI: Notificações -> Telegram.
* Backup SCB na abertura/fechamento de canal (ativar/desativar).
* Resumo financeiro programado (intervalos de hora em hora até 12 horas).
* Comandos sob demanda: `/scb` (backup) e `/balances` (resumo).
* O token do bot vem do @BotFather e o ID do chat do @userinfobot.
* Apenas chat direto; deixar ambos os campos vazios desativa o Telegram.

Chaves de ambiente:

* `NOTIFICATIONS_TG_BOT_TOKEN`
* `NOTIFICATIONS_TG_CHAT_ID`

## Relatórios

Os relatórios diários de roteamento são calculados à meia-noite (horário local) e armazenados no Postgres (mesmo BD/usuário das notificações).

Agendamento:

* O `lightningos-reports.timer` roda o `lightningos-reports.service` às `00:00` do horário local.
* Execução manual: `/opt/lightningos/manager/lightningos-manager reports-run --date YYYY-MM-DD` (o padrão é ontem).
* Preenchimento retroativo (backfill): `/opt/lightningos/manager/lightningos-manager reports-backfill --from YYYY-MM-DD --to YYYY-MM-DD` (padrão máximo de 730 dias; use `--max-days N` para substituir).
* Fixação de fuso horário opcional: defina `REPORTS_TIMEZONE=America/Sao_Paulo` em `/etc/lightningos/secrets.env` para forçar que os relatórios diários, backfill e ao vivo usem o mesmo fuso horário IANA.

Tabela armazenada: `reports_daily`

* `report_date` (DATE, dia local)
* `forward_fee_revenue_sats`
* `forward_fee_revenue_msat`
* `rebalance_fee_cost_sats`
* `rebalance_fee_cost_msat`
* `net_routing_profit_sats`
* `net_routing_profit_msat`
* `forward_count`
* `rebalance_count`
* `routed_volume_sats`
* `routed_volume_msat`
* `onchain_balance_sats`
* `lightning_balance_sats`
* `total_balance_sats`
* `created_at`, `updated_at`

Endpoints da API:

* `GET /api/reports/range?range=d-1|month|3m|6m|12m|all` (mês = últimos 30 dias)
* `GET /api/reports/custom?from=YYYY-MM-DD&to=YYYY-MM-DD` (máx. 730 dias)
* `GET /api/reports/summary?range=...`
* `GET /api/reports/live` (hoje das 00:00 locais → agora, em cache por ~60s)

## Rebalance Center (Central de Rebalanceamento)

O Rebalance Center é um otimizador de liquidez de entrada (local/saída) para o LND. Ele pode executar rebalanceamentos manuais por canal ou varreduras totalmente automáticas que enfileiram rebalanceamentos com base no ROI e em restrições orçamentárias. Um rebalanceamento só prossegue quando a **taxa de saída > taxa do peer**, para que você nunca pague mais do que a taxa cobrada pelo peer sem que haja um spread positivo. Os custos são rastreados através das notificações (fee msat) e agregados ao custo em tempo real + gastos diários automáticos/manuais.

Comportamento principal:

* Rebalanceamentos manuais ignoram o orçamento diário e podem ser iniciados por canal.
* Rebalanceamentos automáticos respeitam o orçamento diário e visam apenas os canais explicitamente marcados como `Auto`.
* Canais de origem são selecionados entre aqueles que têm liquidez local suficiente e que não estão excluídos; um canal preenchido por um rebalanceamento se torna **protegido** (protected) e não pode ser usado como origem até que as regras de pagamento (payback) o liberem.
* Os alvos são escolhidos quando o déficit de liquidez de saída excede a zona morta (deadband) e o spread de taxa é positivo; a estimativa do ROI usa os últimos 7 dias de receita de roteamento em comparação com o custo estimado do rebalanceamento.
* Os alvos automáticos são classificados pela **pontuação econômica (economic score)** = (ganho esperado − custo estimado), assim canais com margens mais altas são priorizados.
* Um **mecanismo de proteção de lucro (profit guardrail)** previne os enfileiramentos automáticos quando o ganho esperado é menor que o custo estimado (quando ambos são conhecidos). Se o ROI for indeterminado (custo = 0 com spread positivo), o modo automático ainda é permitido.
* A seleção de origem é ponderada pelo histórico do par: pares recentes e bem-sucedidos com taxas mais baixas são priorizados, enquanto falhas recentes são despriorizadas.
* A visão geral mostra **Última varredura (Last scan)** no horário local e o status da varredura (ex: sem origens, sem candidatos, orçamento esgotado), além de telemetria econômica (melhor pontuação, ignorados pela proteção de lucro) e detalhes opcionais de "pular" (skip).
* Os rebalanceamentos manuais podem opcionalmente ter **auto-reinício** (alternador por canal) com um tempo de recarga de 60s até que o objetivo seja alcançado.
* A **pré-sondagem de rota (pre-probing)** roda antes do envio, procurando pelo maior valor viável na rota.

Workbench de Canais:

* Define a porcentagem de saída alvo por canal.
* Use a opção `Auto` para permitir que o modo automático rebalanceie o canal.
* Use o ícone de reinício (restart) para auto-reiniciar rebalanceamentos manuais no canal.
* Use a opção `Exclude source` para bloquear que o canal seja usado como origem.
* Alternador de classificação: **Economic** (baseado em pontuação) ou **Emptiest** (menor % local primeiro).

Código de cores (linhas de canais):

* Fundo verde = origem elegível (pode financiar rebalanceamentos).
* Fundo vermelho = alvo elegível (habilitado no automático e precisando de liquidez de saída).
* Fundo âmbar = alvo em potencial (precisa de liquidez de saída mas não está habilitado no automático).

Parâmetros de configuração:

* Configurações apenas automáticas: `Enable auto rebalance`, `Scan interval (sec)`, `Daily budget (% of revenue)`.
* `Enable auto rebalance`: ativa ou desativa a varredura automática.
* `Scan interval (sec)`: a frequência da varredura automática.
* `Daily budget (% of revenue)`: porcentagem da receita de roteamento das últimas 24h alocada para rebalanceamentos automáticos.
* `Deadband (%)`: déficit mínimo de liquidez de saída antes que um canal se torne um alvo.
* `Minimum local for source (%)`: liquidez local mínima necessária para um canal ser uma origem.
* `Economic ratio`: fração da taxa de saída do canal alvo (base+ppm) usada como teto máximo de taxa.
* `Econ ratio max (ppm)`: teto opcional para o limite de taxa ao usar o índice econômico (0 = sem limite).
* `Fee limit (ppm)`: substitui o índice econômico por um limite fixo máximo de ppm (0 = desativado).
* `Subtract source fees`: reduz o orçamento de taxas usando taxas de origem estimadas (mais conservador).
* `ROI minimum`: ROI estimado mínimo (receita de 7 dias / custo estimado) para enfileirar tarefas automáticas.
* `Max concurrent`: número máximo de rebalanceamentos ocorrendo ao mesmo tempo.
* `Minimum (sats)`: menor valor de rebalanceamento em tentativas padrão (a sondagem pode ir abaixo para capturar uma rota válida).
* `Maximum (sats)`: limite superior para o tamanho do rebalanceamento (0 = ilimitado).
* `Fee ladder steps`: número de tetos de taxa para tentar, da mais baixa para a mais alta, antes de desistir.
* `Amount probe steps`: número de sondagens de montante (amount), do maior para o menor, quando ocorre uma falha temporária no último salto (last-hop).
* `Fail tolerance (ppm)`: a sondagem para quando a diferença (delta) entre montantes estiver abaixo deste limite.
* `Adaptive amount probing`: limita a próxima tentativa baseada no último montante bem-sucedido.
* `Attempt timeout (sec)`: tempo máximo por tentativa antes de passar para a próxima taxa/montante.
* `Rebalance timeout (sec)`: tempo máximo de execução por tarefa de rebalanceamento (automático ou manual).
* `Mission control half-life (sec)`: tempo de decaimento para falhas de controle de missão (mission control) (menor = esquece mais rápido, 0 = padrão do LND).
* `Payback policy`: três modos que podem ser habilitados juntos.
* `Release by payback`: desbloqueia a liquidez protegida quando a receita de roteamento pagar o custo do rebalanceamento.
* `Release by time`: desbloqueia após decorridos `Unlock days` (dias para desbloqueio) desde o último rebalanceamento.
* `Critical mode`: desbloqueia uma fração quando as origens são escassas para varreduras repetidas.
* `Unlock days`: número de dias antes do desbloqueio baseado em tempo.
* `Critical release (%)`: porcentagem de liquidez protegida liberada por ciclo crítico.
* `Critical cycles`: varreduras consecutivas com fontes baixas antes que a liberação crítica dispare.
* `Critical min sources`: mínimo de canais de origem elegíveis necessários para evitar o modo crítico.
* `Critical min available sats`: liquidez total mínima de origem necessária para evitar o modo crítico.

## Lightning Ops: Autofee

O Autofee ajusta automaticamente as **taxas de saída (outbound fees)** para maximizar **lucros primeiro** e **movimentação depois**. Ele utiliza o histórico de roteamentos e rebalanceamentos locais (notificações do Postgres) e métricas opcionais da Amboss para semear preços, para então aplicar barreiras (guardrails), tempos de recarga (cooldowns) e limites (caps), de forma que as atualizações sejam seguras e explicáveis.

Parâmetros da UI:

* `Enable autofee`: liga/desliga geral.
* `Profile`: Conservador / Moderado / Agressivo (ajusta limites internos).
* `Lookback window (days)`: 5 a 21 dias para estatísticas.
* `Run interval (hours)`: mínimo de 1 hora.
* `Cooldown up / down (hours)`: tempo mínimo entre aumentos / diminuições de taxas.
* `Min fee (ppm)` e `Max fee (ppm)`: limitações fixas (mínimo pode ser `0`).
* `Rebalance cost mode`: `Per-channel` (Por canal), `Global`, ou `Blend` (Misturado) (controla a âncora de custo usada nos pisos/margens).
* `Amboss fee reference`: fonte opcional de semente (seed); exige token da API.
* `Inbound passive rebalance`: usa o desconto de entrada para canais de "ralo" (sink channels).
* `Discovery mode`: redução mais rápida de taxas para canais inativos ou com saída alta.
* `Explorer mode`: ciclos de exploração temporários; podem pular o cooldown em movimentos de queda.
* `Revenue floor`: mantém um piso (floor) mínimo para canais de alta performance.
* `Circuit breaker`: reduz passos se a demanda cair após aumentos recentes.
* `Extreme drain`: acelera o aumento de taxas quando o canal é drenado cronicamente.
* `Super source` + base fee: aumenta a taxa base quando o canal é classificado como uma super fonte.
* `HTLC signal integration`: ativa o feedback de falha de HTLC do Gerenciador HTLC.
* `HTLC mode`: `observe_only` (apenas telemetria/tags), `policy_only` (apenas efeitos ao lado da política), `full` (efeitos na política + liquidez, padrão).

Comportamento do sinal HTLC:

* A janela do sinal é alinhada com a cadência do Autofee: `max(run_interval, 60m)`.
* Os limites mínimos de falha/amostra são dimensionados pela janela HTLC ativa e calibrados pelo tamanho da liquidez + classe do nó.
* A linha de resumo inclui: `htlc_liq_hot`, `htlc_policy_hot`, `htlc_low_sample`, `htlc_window`.
* A linha por canal inclui contadores (quando presentes): `htlc<window>m a=<attempts> p=<policy_fails> l=<liquidity_fails>`.

Calibração automática:

* Cada execução calcula a classificação do nó e o status da liquidez para dimensionar limites automaticamente.
* Classes de tamanho do nó (baseadas em capacidade total e contagem de canais):
* `small` (pequeno): < 50M sats ou < 20 canais
* `medium` (médio): < 200M sats ou < 60 canais
* `large` (grande): < 1.5B sats ou < 150 canais
* `extra large` (extra grande): tudo acima disso


* Classes de liquidez (baseadas no índice (ratio) local):
* `drained` (drenado): razão local < 25%
* `balanced` (equilibrado): 25% a 75%
* `full` (cheio): razão local > 75%


* A linha de calibração nos Resultados do Autofee mostra estas classes, além de limites calibrados no `revfloor`.

Linhas dos Resultados do Autofee:

* Cabeçalho: tipo de execução e carimbo de data/hora (timestamp).
* Resumo: contagens para subida/descida/estável (up/down/flat) e motivos de pulos (skips).
* Semente (Seed): uso do Amboss e fallbacks.
* Calibração: tamanho do nó, liquidez e limites calibrados.
* Linhas por canal: decisão, alvo, pisos (floors), margens e tags.
* Filtros de resultados: você pode exibir as últimas N execuções e filtrar opcionalmente por período de tempo local.

Glossário de Tags (Resultados do Autofee):

* `sink`, `source`, `router`, `unknown`: rótulos de classes.
* `discovery`: canal em modo de descoberta.
* `discovery-hard`: queda drástica de descoberta acionada (sem baseline + inativo).
* `explorer`: modo de explorador ativado.
* `cooldown-skip`: recarga de pulo bloqueada em queda devido a modo de explorador.
* `surge+X%`: aumento de pico aplicado.
* `top-rev`: aumento de taxa para o líder de receita (revenue share) aplicado.
* `neg-margin`: proteção de margem negativa aplicada.
* `negm+X%`: aumento de margem negativa extra aplicado.
* `revfloor`: piso de receita aplicado.
* `outrate-floor`: piso de taxa de saída aplicado.
* `peg`: fixado (peg) com base na taxa de saída observada.
* `peg-grace`: peg aplicado dentro da janela de tolerância.
* `peg-demand`: peg aplicado devido a uma forte demanda vs seed.
* `circuit-breaker`: o disjuntor (circuit breaker) reduziu a variação (step).
* `extreme-drain`: caminho de aumento por dreno extremo utilizado.
* `extreme-drain-turbo`: aumento turbo de dreno extremo extra adicionado.
* `cooldown`: a recarga (cooldown) bloqueou uma atualização.
* `cooldown-profit`: cooldown paralisou uma movimentação rentável para baixo.
* `hold-small`: variação menor que o mínimo percentual/delta estipulado.
* `same-ppm`: alvo é igual ao ppm atual.
* `no-down-low`: bloqueio de movimentação para baixo enquanto o canal está extremamente drenado.
* `no-down-neg-margin`: bloqueio de movimentação para baixo enquanto a margem é negativa.
* `sink-floor`: margem de teto extra aplicada a canais que escoam a liquidez (sink channels).
* `trend-up`, `trend-down`, `trend-flat`: dica de tendência de movimentação.
* `stepcap`: limite de variação (step cap) aplicado.
* `stepcap-lock`: trava de limite de variação aplicada.
* `floor-lock`: trava de teto (floor lock) aplicada.
* `global-neg-lock`: trava de margem negativa global aplicada.
* `lock-skip-no-chan-rebal`: pulo da trava porque não há histórico de rebalanceamento de canal.
* `lock-skip-sink-profit`: pulo da trava no caminho de lucratividade do sink.
* `profit-protect-lock`: proteção contra perdas (profit-protection) bloqueando operação aplicada.
* `profit-protect-relax`: relaxamento na proteção de lucros.
* `super-source`: canal classificado como uma super fonte (super source).
* `super-source-like`: canal parecido com router em classe super-source.
* `inb-<n>`: desconto de entrada (inbound) ativado.
* `htlc-policy-hot`: alto sinal em falhas por política no HTLC.
* `htlc-liquidity-hot`: alto sinal em falhas por liquidez no HTLC.
* `htlc-sample-low`: amostragem do HTLC é muito pequena para receber o sinal hot.
* `htlc-neutral-lock`: bloqueio num caminho neutro usado já que ambos os sinais hot da HTLC estão engatados.
* `htlc-liq+X%`: acréscimo de liquidez com tag-hot adicionado.
* `htlc-policy+X%`: acréscimo à política devido à tag-hot adicionada.
* `htlc-liq-nodown`: movimentação impedida porque tem um alto sinal de falhas na HTLC.
* `htlc-policy-nodown`: movimentação em taxa baixa freada com motivo HTLC de tag-hot.
* `htlc-neutral-nodown`: impedimento em queda gerado por um modelo combinado HTLC.
* `htlc-step-boost`: acréscimo em pulos devido ao sinal HTLC hot presente.
* `seed:amboss`: Amboss seed utilizada.
* `seed:amboss-missing`: Faltando o Token do Amboss.
* `seed:amboss-empty`: O Amboss retornou dados vazios.
* `seed:amboss-error`: Ocorreu erro ao usar o Amboss.
* `seed:med`: A mediana do Amboss foi misturada na análise.
* `seed:vol-<n>%`: Penalidade por volatilidade na taxa aplicada.
* `seed:ratio<factor>`: Relação entre a saída/entrada gerou ajuste à taxa aplicada.
* `seed:outrate`: A seed partiu da nossa taxa de saída mais recente.
* `seed:mem`: Semente extraída através de dados via memória.
* `seed:default`: Método fallback tradicional adotado (seed padrão).
* `seed:guard`: A segurança pulou ao aplicar a proteção à taxa do seed.
* `seed:p95cap`: Proteção à semente foi baseada nas métricas do cap de limite de percentis 95 da ferramenta do Amboss.
* `seed:absmax`: Um corte absoluto contra seed alta foi imposto.

## Terminal Web (opcional)

O LightningOS Light pode expor um terminal web protegido usando GoTTY.

O instalador ativa o terminal automaticamente e gera uma credencial quando ela está faltando.
Você pode revisar ou alterar no `/etc/lightningos/secrets.env`:

* `TERMINAL_ENABLED=1`
* `TERMINAL_CREDENTIAL=user:pass`
* `TERMINAL_ALLOW_WRITE=0` (defina como `1` para permitir a digitação)
* `TERMINAL_PORT=7681` (opcional)
* `TERMINAL_WS_ORIGIN=^https://.*:8443$` (opcional, o padrão permite todas as origens)

Inicie (ou reinicie) o serviço:

```bash
sudo systemctl enable --now lightningos-terminal

```

A página do Terminal exibe a senha atual e um botão para copiar.

## Notas de segurança

* A frase semente (seed phrase) nunca é armazenada. Ela é exibida uma vez no assistente.
* Credenciais RPC são armazenadas apenas em `/etc/lightningos/secrets.env` (root:lightningos, `chmod 660`).
* Por padrão, a API/UI se conecta a `0.0.0.0` para acesso via LAN. Caso deseje que o acesso seja feito apenas do localhost, reconfigure `server.host: "127.0.0.1"` no arquivo `/etc/lightningos/config.yaml`.

## Solução de problemas (Troubleshooting)

Caso `https://<IP_DA_LAN_DO_SERVIDOR>:8443` não esteja acessível:

```bash
systemctl status lightningos-manager --no-pager
journalctl -u lightningos-manager -n 200 --no-pager
ss -ltn | grep :8443

```

### App Store (LNDg, Peerswap, Elements, Bitcoin Core)

* O LNDg é executado no Docker e escuta em `http://<IP_DA_LAN_DO_SERVIDOR>:8889`.
* O Peerswap instala o `peerswapd` + `psweb` (UI em `http://<IP_DA_LAN_DO_SERVIDOR>:1984`) e necessita do Elements.
* O Elements roda como um serviço nativo (Nó Liquid Elements, RPC na porta `127.0.0.1:7041`).
* O Bitcoin Core executa via Docker com os dados alocados dentro de `/data/bitcoin`.

Notas sobre o LNDg:

* A aba de logs do LNDg examina `/var/log/lndg-controller.log` na configuração que funciona dentro do container. Caso apareça que ela está sem nada, inspecione a rotina: `docker logs lndg-lndg-1`.
* Caso você note uma mensagem dizendo `Is a directory: /var/log/lndg-controller.log`, será preciso descartar o diretório através de `/var/lib/lightningos/apps-data/lndg/data/lndg-controller.log` situado na sua máquina nativa e, após esse processo, inicie o LNDg novamente.
* Note que se o LND recorrer ao Postgres, o aplicativo LNDg pode registrar no log que existe uma falha com relação ao banco `channel.db`. Tal erro não surtirá nenhum efeito danoso na execução do sistema.

## Arquitetura da App Store

* O manipulador de rotina responsável pelos Apps atua em: `internal/server/apps_<app>.go`.
* Todo App vai se registrar através da página: `internal/server/apps_registry.go`.
* Todos os arquivos locais dos componentes são listados na subpasta interna alocada no sistema via `/var/lib/lightningos/apps/<app>` ao passo que os registros duráveis (persistent data) ficam armazenados em `/var/lib/lightningos/apps-data/<app>`.
* O Docker deve ser ativado conforme as circunstâncias estipuladas por qualquer programa que requisite desse tipo de motor. O plano base do instalador, em si, pode durar ileso e continuar atuando longe do alcance do Docker em suas dependências básicas.
* Registros que fazem avaliações com base na sanidade estrutural asseguram o funcionamento ao atuar sobre checagens que exigem IDs singulares e a atribuição de canais de porta singulares à parte.

### Adicionar um novo app

1. Programe e crie a via em `internal/server/apps_<app>.go` provendo sua classe principal dentro da interface estipulada em `appHandler`.
2. Submeta formalmente este app para ser listado no repositório final de `internal/server/apps_registry.go`.
3. Em sequência adicione o cartão para compor o app atrelado com: `ui/src/pages/AppStore.tsx` englobando logo e marca da respectiva pasta contida no repositório final em: `ui/src/assets/apps/`.

### Verificações do App Store

Verifique os registros efetuando testes a nível de sanidade:

```bash
go test ./internal/server -run TestValidateAppRegistry

```

## Desenvolvimento

Confira os requisitos em `DEVELOPMENT.md` para preparar um modelo em base local a fim de conseguir realizar o build inicial da sua construção local do sistema.

## Systemd

Você encontrará esses scripts-modelos contidos em `templates/systemd/`.

## Reconstruir apenas (Apenas rebuild em manager/UI)

Muitas vezes, a necessidade base atua em apenas reverter os processos estritos para se compilar as novas chaves, e isso por si já soluciona sem a necessidade do instalador principal. Use esses modelos neste caso:

Para reprocessar o painel final Manager:

```bash
sudo /usr/local/go/bin/go build -o dist/lightningos-manager ./cmd/lightningos-manager
sudo install -m 0755 dist/lightningos-manager /opt/lightningos/manager/lightningos-manager
sudo systemctl restart lightningos-manager

```

Refazer build focando a Interface ao Usuário da UI (Apenas UI):

```bash
cd ui && sudo npm install && sudo npm run build
cd ..
sudo rm -rf /opt/lightningos/ui/*
sudo cp -a ui/dist/. /opt/lightningos/ui/

```
