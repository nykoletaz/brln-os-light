package lndclient

import (
  "context"
  "crypto/rand"
  "crypto/sha256"
  "encoding/hex"
  "errors"
  "fmt"
  "strings"

  "lightningos-light/lnrpc"
)

const (
  KeysendPreimageRecord uint64 = 5482373484
  KeysendMessageRecord  uint64 = 34349334
)

func (c *Client) SendKeysendMessage(ctx context.Context, pubkeyHex string, amountSat int64, message string) (string, error) {
  trimmed := strings.TrimSpace(pubkeyHex)
  if trimmed == "" {
    return "", errors.New("pubkey required")
  }
  if amountSat <= 0 {
    return "", errors.New("amount must be positive")
  }
  pubkey, err := hex.DecodeString(trimmed)
  if err != nil {
    return "", fmt.Errorf("invalid pubkey hex")
  }
  if len(pubkey) != 33 {
    return "", fmt.Errorf("invalid pubkey length")
  }

  preimage := make([]byte, 32)
  if _, err := rand.Read(preimage); err != nil {
    return "", err
  }
  hash := sha256.Sum256(preimage)

  conn, err := c.dial(ctx, true)
  if err != nil {
    return "", err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)
  _, err = client.SendPaymentSync(ctx, &lnrpc.SendRequest{
    Dest: pubkey,
    Amt: amountSat,
    PaymentHash: hash[:],
    DestCustomRecords: map[uint64][]byte{
      KeysendPreimageRecord: preimage,
      KeysendMessageRecord:  []byte(message),
    },
  })
  if err != nil {
    return "", err
  }

  return hex.EncodeToString(hash[:]), nil
}
