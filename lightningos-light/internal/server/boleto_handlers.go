package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	fswapNodeAPIDefault = "https://api.f-swap.com/node"
	secretsEnvPath      = "/etc/lightningos/secrets.env"
)

// ─── Types ──────────────────────────────────────────────────────────

type boletoQuoteRequest struct {
	Barcode string `json:"barcode"`
}

type boletoQuoteResponse struct {
	Invoice     string  `json:"invoice"`
	PaymentHash string  `json:"paymentHash"`
	AmountBrl   float64 `json:"amountBrl"`
	BaseSats    int64   `json:"baseSats"`
	FeeSats     int64   `json:"feeSats"`
	FeePercent  int     `json:"feePercent"`
	TotalSats   int64   `json:"totalSats"`
	BtcRate     float64 `json:"btcRate"`
	BankName    string  `json:"bankName"`
	BankCode    string  `json:"bankCode"`
	DueDate     *string `json:"dueDate"`
	ExpiresAt   string  `json:"expiresAt"`
}

type boletoStatusResponse struct {
	Status      string  `json:"status"`
	InvoicePaid bool    `json:"invoicePaid"`
	BoletoPaid  bool    `json:"boletoPaid"`
	Protocolo   string  `json:"protocolo,omitempty"`
	AmountBrl   float64 `json:"amountBrl,omitempty"`
	TotalSats   int64   `json:"totalSats,omitempty"`
	Reason      string  `json:"reason,omitempty"`
}

type fswapErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ─── Helpers ────────────────────────────────────────────────────────

func fswapNodeAPIKey() string {
	return strings.TrimSpace(os.Getenv("FSWAP_NODE_API_KEY"))
}

func fswapAPIBase() string {
	if v := strings.TrimSpace(os.Getenv("FSWAP_API_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fswapNodeAPIDefault
}

func fswapRequest(method, path string, body any, apiKey string) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	url := fswapAPIBase() + path
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-Node-Api-Key", apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

func fswapNodeRequest(method, path string, body any) (*http.Response, error) {
	apiKey := fswapNodeAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("FSWAP_NODE_API_KEY not configured")
	}
	return fswapRequest(method, path, body, apiKey)
}

// ─── Handlers ───────────────────────────────────────────────────────

// handleBoletoConfig returns config/availability for boleto payments
func (s *Server) handleBoletoConfig(w http.ResponseWriter, r *http.Request) {
	apiKey := fswapNodeAPIKey()
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":    true,
		"activated":  apiKey != "",
		"feePercent": 6,
		"provider":   "fswap",
	})
}

// handleBoletoQuote proxies to FSwap backend to create a quote
func (s *Server) handleBoletoQuote(w http.ResponseWriter, r *http.Request) {
	if fswapNodeAPIKey() == "" {
		writeError(w, http.StatusServiceUnavailable, "Pagamento de boletos não configurado")
		return
	}

	var payload boletoQuoteRequest
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	payload.Barcode = strings.TrimSpace(payload.Barcode)
	if len(payload.Barcode) < 44 {
		writeError(w, http.StatusBadRequest, "Código de barras deve ter pelo menos 44 dígitos")
		return
	}

	resp, err := fswapNodeRequest("POST", "/boleto/quote", payload)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao conectar com serviço de pagamentos: "+err.Error())
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao ler resposta")
		return
	}

	// Forward the status code and body from FSwap
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}

// handleBoletoStatus proxies to FSwap backend to check payment status
func (s *Server) handleBoletoStatus(w http.ResponseWriter, r *http.Request) {
	if fswapNodeAPIKey() == "" {
		writeError(w, http.StatusServiceUnavailable, "Pagamento de boletos não configurado")
		return
	}

	paymentHash := chi.URLParam(r, "paymentHash")
	if paymentHash == "" || len(paymentHash) < 32 {
		writeError(w, http.StatusBadRequest, "paymentHash inválido")
		return
	}

	resp, err := fswapNodeRequest("GET", "/boleto/status/"+paymentHash, nil)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao conectar com serviço de pagamentos: "+err.Error())
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao ler resposta")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}

// ─── Activation handlers ────────────────────────────────────────────

// handleBoletoActivate creates an activation invoice via FSwap.
// The node pubkey is fetched from LND automatically.
func (s *Server) handleBoletoActivate(w http.ResponseWriter, r *http.Request) {
	// Don't allow if already activated
	if fswapNodeAPIKey() != "" {
		writeError(w, http.StatusConflict, "Node já ativado")
		return
	}

	// Get node pubkey from LND
	ctx := r.Context()
	status, err := s.lnd.GetStatus(ctx)
	if err != nil || status.Pubkey == "" {
		writeError(w, http.StatusServiceUnavailable, "Não foi possível obter pubkey do node LND")
		return
	}

	// Call FSwap /node/activate (no API key needed)
	resp, err := fswapRequest("POST", "/activate", map[string]string{
		"nodePubkey": status.Pubkey,
	}, "")
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao conectar com FSwap: "+err.Error())
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao ler resposta")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}

// handleBoletoActivateStatus polls activation status. When completed,
// saves the API key to secrets.env and reloads environment.
func (s *Server) handleBoletoActivateStatus(w http.ResponseWriter, r *http.Request) {
	paymentHash := chi.URLParam(r, "paymentHash")
	if paymentHash == "" || len(paymentHash) < 32 {
		writeError(w, http.StatusBadRequest, "paymentHash inválido")
		return
	}

	// Call FSwap /node/activate/status/:paymentHash (no API key needed)
	resp, err := fswapRequest("GET", "/activate/status/"+paymentHash, nil, "")
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao conectar com FSwap: "+err.Error())
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Erro ao ler resposta")
		return
	}

	// If activation completed, save the API key to secrets.env + process env
	if resp.StatusCode == 200 {
		var result struct {
			Status string `json:"status"`
			APIKey string `json:"apiKey"`
		}
		if err := json.Unmarshal(data, &result); err == nil && result.Status == "completed" && result.APIKey != "" {
			if saveErr := writeEnvFileValue(secretsEnvPath, "FSWAP_NODE_API_KEY", result.APIKey); saveErr == nil {
				os.Setenv("FSWAP_NODE_API_KEY", result.APIKey)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}
