module lightningos-light

go 1.22

require (
  github.com/go-chi/chi/v5 v5.0.10
  github.com/go-chi/cors v1.2.1
  github.com/jackc/pgx/v5 v5.5.5
  github.com/lightningnetwork/lnd v0.18.4-beta
  gopkg.in/yaml.v3 v3.0.1
  google.golang.org/grpc v1.59.0
  google.golang.org/protobuf v1.33.0
)

replace google.golang.org/protobuf => github.com/lightninglabs/protobuf-go-hex-display v1.30.0-hex-display
