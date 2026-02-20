package server

import (
  "context"
  "errors"
  "log"
  "net/url"
  "strings"

  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgxpool"
)

const (
  shortcutIconTypeEmoji = "emoji"
  shortcutIconTypeImage = "image"
)

var (
  ErrShortcutInvalidURL = errors.New("invalid url")
  ErrShortcutInvalidEmoji = errors.New("invalid emoji")
  ErrShortcutExists = errors.New("shortcut already exists")
  ErrShortcutNotFound = errors.New("shortcut not found")
  ErrShortcutProtected = errors.New("default shortcut cannot be removed")
)

type Shortcut struct {
  ID int64 `json:"id"`
  URL string `json:"url"`
  Name string `json:"name"`
  Description string `json:"description"`
  IconType string `json:"icon_type"`
  IconValue string `json:"icon_value"`
  SortOrder int `json:"sort_order"`
  Protected bool `json:"protected"`
}

type shortcutSeed struct {
  URL string
  Name string
  Description string
  IconType string
  IconValue string
  SortOrder int
}

var defaultShortcuts = []shortcutSeed{
  {
    URL: "https://services.br-ln.com",
    Name: "BR\u26A1LN Services",
    Description: "Clube",
    IconType: shortcutIconTypeImage,
    IconValue: "/shortcuts/brln-club.svg",
    SortOrder: 10,
  },
  {
    URL: "https://wallet.br-ln.com",
    Name: "BR\u26A1LN Wallet",
    Description: "",
    IconType: shortcutIconTypeImage,
    IconValue: "/shortcuts/brln-club.svg",
    SortOrder: 20,
  },
  {
    URL: "https://pay.br-ln.com",
    Name: "BRL\u26A1N Lightning Address",
    Description: "",
    IconType: shortcutIconTypeImage,
    IconValue: "/shortcuts/brln-club.svg",
    SortOrder: 30,
  },
}

type ShortcutsService struct {
  db *pgxpool.Pool
  logger *log.Logger
}

func NewShortcutsService(db *pgxpool.Pool, logger *log.Logger) *ShortcutsService {
  return &ShortcutsService{
    db: db,
    logger: logger,
  }
}

func (s *ShortcutsService) EnsureSchema(ctx context.Context) error {
  if s == nil || s.db == nil {
    return errors.New("shortcuts db unavailable")
  }

  _, err := s.db.Exec(ctx, `
create table if not exists ui_shortcuts (
  id bigserial primary key,
  url text not null unique,
  name text not null,
  description text not null default '',
  icon_type text not null check (icon_type in ('emoji','image')),
  icon_value text not null,
  sort_order integer not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists ui_shortcuts_sort_idx on ui_shortcuts (sort_order asc, id asc);
`)
  if err != nil {
    return err
  }

  var count int64
  if err := s.db.QueryRow(ctx, `select count(*) from ui_shortcuts`).Scan(&count); err != nil {
    return err
  }
  if count > 0 {
    return nil
  }

  for _, seed := range defaultShortcuts {
    _, err := s.db.Exec(ctx, `
insert into ui_shortcuts (url, name, description, icon_type, icon_value, sort_order)
values ($1, $2, $3, $4, $5, $6)
`, seed.URL, seed.Name, seed.Description, seed.IconType, seed.IconValue, seed.SortOrder)
    if err != nil {
      return err
    }
  }
  return nil
}

func (s *ShortcutsService) List(ctx context.Context) ([]Shortcut, error) {
  if s == nil || s.db == nil {
    return nil, errors.New("shortcuts db unavailable")
  }
  rows, err := s.db.Query(ctx, `
select id, url, name, description, icon_type, icon_value, sort_order
from ui_shortcuts
order by sort_order asc, id asc
`)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  items := []Shortcut{}
  for rows.Next() {
    var item Shortcut
    if err := rows.Scan(
      &item.ID,
      &item.URL,
      &item.Name,
      &item.Description,
      &item.IconType,
      &item.IconValue,
      &item.SortOrder,
    ); err != nil {
      return nil, err
    }
    item.Protected = isProtectedShortcutIconType(item.IconType)
    items = append(items, item)
  }
  return items, rows.Err()
}

func (s *ShortcutsService) Create(ctx context.Context, rawURL, rawEmoji string) (Shortcut, error) {
  if s == nil || s.db == nil {
    return Shortcut{}, errors.New("shortcuts db unavailable")
  }

  normalizedURL, err := normalizeShortcutURL(rawURL)
  if err != nil {
    return Shortcut{}, err
  }
  emoji, err := normalizeShortcutEmoji(rawEmoji)
  if err != nil {
    return Shortcut{}, err
  }

  sortOrder := 10
  if err := s.db.QueryRow(ctx, `select coalesce(max(sort_order), 0) + 10 from ui_shortcuts`).Scan(&sortOrder); err != nil {
    return Shortcut{}, err
  }

  item := Shortcut{
    URL: normalizedURL,
    Name: shortcutNameFromURL(normalizedURL),
    Description: "",
    IconType: shortcutIconTypeEmoji,
    IconValue: emoji,
    SortOrder: sortOrder,
    Protected: false,
  }

  err = s.db.QueryRow(ctx, `
insert into ui_shortcuts (url, name, description, icon_type, icon_value, sort_order)
values ($1, $2, $3, $4, $5, $6)
on conflict (url) do nothing
returning id
`, item.URL, item.Name, item.Description, item.IconType, item.IconValue, item.SortOrder).Scan(&item.ID)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return Shortcut{}, ErrShortcutExists
    }
    return Shortcut{}, err
  }
  return item, nil
}

func (s *ShortcutsService) Delete(ctx context.Context, id int64) error {
  if s == nil || s.db == nil {
    return errors.New("shortcuts db unavailable")
  }
  if id <= 0 {
    return ErrShortcutNotFound
  }
  var iconType string
  if err := s.db.QueryRow(ctx, `select icon_type from ui_shortcuts where id = $1`, id).Scan(&iconType); err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return ErrShortcutNotFound
    }
    return err
  }
  if isProtectedShortcutIconType(iconType) {
    return ErrShortcutProtected
  }
  result, err := s.db.Exec(ctx, `delete from ui_shortcuts where id = $1`, id)
  if err != nil {
    return err
  }
  if result.RowsAffected() == 0 {
    return ErrShortcutNotFound
  }
  return nil
}

func normalizeShortcutURL(raw string) (string, error) {
  value := strings.TrimSpace(raw)
  if value == "" {
    return "", ErrShortcutInvalidURL
  }
  if !strings.Contains(value, "://") {
    value = "https://" + value
  }
  parsed, err := url.Parse(value)
  if err != nil {
    return "", ErrShortcutInvalidURL
  }
  scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
  if scheme != "http" && scheme != "https" {
    return "", ErrShortcutInvalidURL
  }
  if strings.TrimSpace(parsed.Hostname()) == "" {
    return "", ErrShortcutInvalidURL
  }
  parsed.Fragment = ""
  return parsed.String(), nil
}

func normalizeShortcutEmoji(raw string) (string, error) {
  value := strings.TrimSpace(raw)
  if value == "" {
    return "", ErrShortcutInvalidEmoji
  }
  runeCount := len([]rune(value))
  if runeCount == 0 || runeCount > 8 {
    return "", ErrShortcutInvalidEmoji
  }
  return value, nil
}

func shortcutNameFromURL(rawURL string) string {
  parsed, err := url.Parse(rawURL)
  if err != nil {
    return "External App"
  }
  host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
  host = strings.TrimPrefix(host, "www.")
  if host == "" {
    return "External App"
  }
  if len(host) > 64 {
    return host[:64]
  }
  return host
}

func isProtectedShortcutIconType(iconType string) bool {
  return strings.EqualFold(strings.TrimSpace(iconType), shortcutIconTypeImage)
}
