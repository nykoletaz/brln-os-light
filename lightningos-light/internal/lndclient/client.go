package lndclient

import (
  "context"
  "crypto/x509"
  "encoding/hex"
  "fmt"
  "log"
  "os"
  "strings"
  "time"

  "lightningos-light/internal/config"
  "lightningos-light/lnrpc"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
)

type Client struct {
  cfg *config.Config
  logger *log.Logger
}

func New(cfg *config.Config, logger *log.Logger) *Client {
  return &Client{cfg: cfg, logger: logger}
}

type macaroonCredential struct {
  macaroon string
}

func (m macaroonCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
  return map[string]string{"macaroon": m.macaroon}, nil
}

func (m macaroonCredential) RequireTransportSecurity() bool {
  return true
}

func (c *Client) dial(ctx context.Context, withMacaroon bool) (*grpc.ClientConn, error) {
  tlsCert, err := os.ReadFile(c.cfg.LND.TLSCertPath)
  if err != nil {
    return nil, err
  }
  certPool := x509.NewCertPool()
  if ok := certPool.AppendCertsFromPEM(tlsCert); !ok {
    return nil, fmt.Errorf("failed to parse LND TLS cert")
  }

  creds := credentials.NewClientTLSFromCert(certPool, "")
  opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

  if withMacaroon {
    macBytes, err := os.ReadFile(c.cfg.LND.AdminMacaroonPath)
    if err != nil {
      return nil, err
    }
    macCred := macaroonCredential{hex.EncodeToString(macBytes)}
    opts = append(opts, grpc.WithPerRPCCredentials(macCred))
  }

  return grpc.DialContext(ctx, c.cfg.LND.GRPCHost, opts...)
}

func (c *Client) GetStatus(ctx context.Context) (Status, error) {
  conn, err := c.dial(ctx, true)
  if err != nil {
    return Status{WalletState: "unknown"}, err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)

  info, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
  if err != nil {
    if isWalletLocked(err) {
      return Status{WalletState: "locked"}, nil
    }
    return Status{WalletState: "unknown"}, err
  }

  status := Status{
    ServiceActive: true,
    WalletState: "unlocked",
    SyncedToChain: info.SyncedToChain,
    SyncedToGraph: info.SyncedToGraph,
    BlockHeight: int64(info.BlockHeight),
  }

  channels, err := client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
  if err == nil {
    active := 0
    inactive := 0
    for _, ch := range channels.Channels {
      if ch.Active {
        active++
      } else {
        inactive++
      }
    }
    status.ChannelsActive = active
    status.ChannelsInactive = inactive
  }

  wallet, err := client.WalletBalance(ctx, &lnrpc.WalletBalanceRequest{})
  if err == nil {
    status.OnchainSat = wallet.TotalBalance
  }

  channelBal, err := client.ChannelBalance(ctx, &lnrpc.ChannelBalanceRequest{})
  if err == nil {
    status.LightningSat = channelBal.Balance
  }

  return status, nil
}

func (c *Client) GenSeed(ctx context.Context, walletPassword string) ([]string, error) {
  conn, err := c.dial(ctx, false)
  if err != nil {
    return nil, err
  }
  defer conn.Close()

  client := lnrpc.NewWalletUnlockerClient(conn)

  resp, err := client.GenSeed(ctx, &lnrpc.GenSeedRequest{})
  if err != nil {
    return nil, err
  }

  return resp.CipherSeedMnemonic, nil
}

func (c *Client) InitWallet(ctx context.Context, walletPassword string, seedWords []string) error {
  conn, err := c.dial(ctx, false)
  if err != nil {
    return err
  }
  defer conn.Close()

  client := lnrpc.NewWalletUnlockerClient(conn)

  _, err = client.InitWallet(ctx, &lnrpc.InitWalletRequest{
    WalletPassword: []byte(walletPassword),
    CipherSeedMnemonic: seedWords,
  })
  return err
}

func (c *Client) UnlockWallet(ctx context.Context, walletPassword string) error {
  conn, err := c.dial(ctx, false)
  if err != nil {
    return err
  }
  defer conn.Close()

  client := lnrpc.NewWalletUnlockerClient(conn)

  _, err = client.UnlockWallet(ctx, &lnrpc.UnlockWalletRequest{WalletPassword: []byte(walletPassword)})
  return err
}

func (c *Client) CreateInvoice(ctx context.Context, amountSat int64, memo string) (string, error) {
  conn, err := c.dial(ctx, true)
  if err != nil {
    return "", err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)

  resp, err := client.AddInvoice(ctx, &lnrpc.Invoice{
    Memo: memo,
    Value: amountSat,
  })
  if err != nil {
    return "", err
  }

  return resp.PaymentRequest, nil
}

func (c *Client) PayInvoice(ctx context.Context, paymentRequest string) error {
  conn, err := c.dial(ctx, true)
  if err != nil {
    return err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)

  _, err = client.SendPaymentSync(ctx, &lnrpc.SendRequest{PaymentRequest: paymentRequest})
  return err
}

func (c *Client) ListRecent(ctx context.Context, limit int) ([]RecentActivity, error) {
  if limit <= 0 {
    limit = 20
  }

  conn, err := c.dial(ctx, true)
  if err != nil {
    return nil, err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)

  invoices, invErr := client.ListInvoices(ctx, &lnrpc.ListInvoiceRequest{Reversed: true, NumMaxInvoices: uint64(limit)})
  payments, payErr := client.ListPayments(ctx, &lnrpc.ListPaymentsRequest{IncludeIncomplete: true, MaxPayments: uint64(limit)})

  var items []RecentActivity
  if invErr == nil {
    for _, inv := range invoices.Invoices {
      items = append(items, RecentActivity{
        Type: "invoice",
        AmountSat: inv.Value,
        Memo: inv.Memo,
        Timestamp: time.Unix(inv.CreationDate, 0).UTC(),
        Status: inv.State.String(),
      })
    }
  }
  if payErr == nil {
    for _, pay := range payments.Payments {
      items = append(items, RecentActivity{
        Type: "payment",
        AmountSat: pay.ValueSat,
        Memo: pay.PaymentRequest,
        Timestamp: time.Unix(pay.CreationDate, 0).UTC(),
        Status: pay.Status.String(),
      })
    }
  }

  return items, nil
}

func isWalletLocked(err error) bool {
  msg := strings.ToLower(err.Error())
  return strings.Contains(msg, "wallet locked") || strings.Contains(msg, "unlock")
}

type Status struct {
  ServiceActive bool
  WalletState string
  SyncedToChain bool
  SyncedToGraph bool
  BlockHeight int64
  ChannelsActive int
  ChannelsInactive int
  OnchainSat int64
  LightningSat int64
}

type RecentActivity struct {
  Type string `json:"type"`
  AmountSat int64 `json:"amount_sat"`
  Memo string `json:"memo"`
  Timestamp time.Time `json:"timestamp"`
  Status string `json:"status"`
}
