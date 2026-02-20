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

const fswapNodeAPIDefault = "https://api.f-swap.com/node"

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

func fswapNodeRequest(method, path string, body any) (*http.Response, error) {
	apiKey := fswapNodeAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("FSWAP_NODE_API_KEY not configured")
	}

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
	req.Header.Set("X-Node-Api-Key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

// ─── Handlers ───────────────────────────────────────────────────────

// handleBoletoConfig returns config/availability for boleto payments
func (s *Server) handleBoletoConfig(w http.ResponseWriter, r *http.Request) {
	apiKey := fswapNodeAPIKey()
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":    apiKey != "",
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
