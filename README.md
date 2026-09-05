# Feature-Flag-Service

Ein thread-sicherer Feature-Flag-Service als REST-API in Go. Flags lassen sich anlegen, auflisten, einzeln abrufen, ändern, löschen und für einen Nutzer deterministisch evaluieren. Der Service nutzt ausschließlich die Standardbibliothek (`net/http`) und hält alle Daten in einem synchronisierten In-Memory-Store.

## Tech-Stack

- Go 1.22+
- `net/http` aus der Standardbibliothek (kein externes Web-Framework)
- Thread-sicherer In-Memory-Store mit `sync.Mutex`
- Tests mit `httptest` und dem `testing`-Paket der Standardbibliothek

## Installation & Start

Es sind keine externen Abhängigkeiten oder Datenbanken nötig — eine Go-Installation (1.22+) genügt.

```bash
go run .
```

Der Server lauscht standardmäßig auf Port **8080**. Über die Umgebungsvariable `PORT` lässt sich ein anderer Port wählen:

```bash
PORT=9000 go run .
```

## Endpunkte

| Methode | Pfad                    | Beschreibung                          |
|---------|-------------------------|---------------------------------------|
| GET     | `/healthz`              | Health-Check → `{"status":"ok"}` (200)|
| POST    | `/flags`                | Flag anlegen                          |
| GET     | `/flags`                | Alle Flags auflisten                  |
| GET     | `/flags/{key}`          | Einzelnes Flag abrufen                |
| PUT     | `/flags/{key}`          | Flag ändern                           |
| DELETE  | `/flags/{key}`          | Flag löschen                          |
| GET     | `/flags/{key}/evaluate` | Flag für einen Nutzer evaluieren      |

Alle Routen sind registriert. Aktuell ist `GET /healthz` der einzige vollständig implementierte Endpunkt; die übrigen antworten bis zur Fertigstellung der zugehörigen Tickets mit `501 Not Implemented`.

Eine nicht unterstützte HTTP-Methode auf einem bekannten Pfad beantwortet der Service mit `405` als JSON-Fehlerobjekt inklusive `Allow`-Header. Jede JSON-Antwort setzt den Header `X-Content-Type-Options: nosniff`; Fehlerantworten mit Status `500` enthalten ausschließlich `{"error":"internal server error"}`.

## Tests

```bash
go test ./...
```
