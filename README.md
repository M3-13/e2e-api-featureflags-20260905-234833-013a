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

## Datenschutz & Betrieb

### Verarbeitungszweck

Der Dienst verarbeitet die Nutzerkennung `user` ausschließlich zur einmaligen, deterministischen Evaluierung eines Feature-Flags pro Nutzer. Eine weitergehende Verwendung — insbesondere Profiling, Tracking oder Weitergabe an Dritte — findet nicht statt.

### Datenkategorien

- **Nutzerkennung** (`user`-Parameter der Evaluierungsroute) — wird nur transient zur Berechnung des Rollout-Ergebnisses genutzt.
- **Flag-Beschreibungstexte** (`description`) — vom Betreiber hinterlegte Metadaten zu den einzelnen Feature-Flags.

### Rechtsgrundlage

Die Verarbeitung erfolgt je nach Einsatz nach **Art. 6 Abs. 1 lit. b DSGVO** (Erfüllung eines Vertrags bzw. vorvertraglicher Maßnahmen) oder nach **Art. 6 Abs. 1 lit. f DSGVO** (berechtigtes Interesse des Verantwortlichen am gesteuerten Rollout von Produktfunktionen).

### Speicherung

Die Nutzerkennung `user` wird **nicht persistiert** — sie verlässt den Prozess nicht und wird weder geloggt noch dauerhaft abgelegt. Flag-Definitionen werden ausschließlich im **flüchtigen In-Memory-Store** gehalten und gehen bei einem Neustart des Dienstes verloren.

### Technische und organisatorische Maßnahmen (TOMs)

- **TLS-Transport** bzw. **TLS-Terminierung** durch einen dem Dienst vorgelagerten Reverse-Proxy.
- **Zugriffsschutz**: Schreibzugriffe (`POST`/`PUT`/`DELETE`) sind über einen `X-API-Key`-Header (`ADMIN_API_KEY`) geschützt.
- **Server-Timeouts** (Read-/Write-/Idle-/Header-Timeout) gegen langsame Verbindungen.
- **Logging ohne Query-Strings**: Der `user`-Parameter erscheint nicht in den Logs.
- **Description-Limit** von 500 Zeichen pro Flag.
- **Body-Limit** von 1 MiB für `POST`/`PUT`-Anfragen.
- **`X-Content-Type-Options: nosniff`** auf allen JSON-Antworten.

### Betroffenenrechte

Betroffene Personen können ihre Rechte nach DSGVO (u. a. Auskunft, Berichtigung, Löschung, Einschränkung der Verarbeitung, Datenübertragbarkeit, Widerspruch) gegenüber dem jeweiligen **Verantwortlichen** geltend machen.

## Sicherheit

### Unterstützte Go-Version

Der Dienst unterstützt **Go 1.22 oder neuer**. Ältere Versionen werden nicht unterstützt.

### Update- und Patch-Prozess

Sicherheitsrelevante Aktualisierungen der Go-Laufzeit und des Quellcodes werden über den regulären Release-Prozess eingespielt. Da ausschließlich die Go-Standardbibliothek verwendet wird, ist die Go-Laufzeit selbst die einzige zu aktualisierende Abhängigkeit.

### Sicherheitsmeldungen

Sicherheitslücken bitte vertraulich an die Sicherheitskontaktadresse des Betreibers melden (z. B. `security@example.com`, vom Betreiber zu ersetzen).

### Dokumentierte Sicherheitseigenschaften

- **Body-Limit** von 1 MiB für `POST`/`PUT`-Anfragen.
- **Eingabevalidierung** (Flag-`key` 1–128 Zeichen, nur `[A-Za-z0-9._-]`; `rollout_percent` 0–100; `user` 1–128 Zeichen).
- **Keine Query-Strings in Logs** (Datenschutz der `user`-Kennung).
- **`X-Content-Type-Options: nosniff`** auf allen JSON-Antworten.
- **500-Maskierung**: Fehlerantworten mit Status 500 enthalten ausschließlich `{"error":"internal server error"}`.

### Betriebsvorgaben

- **Schreibzugriffe schützen**: `POST`/`PUT`/`DELETE`-Anfragen sind über den Header `X-API-Key` mit `ADMIN_API_KEY` abzusichern.
- **TLS-Terminierung**: TLS ist über einen dem Dienst vorgelagerten Reverse-Proxy zu terminieren.

## Tests

```bash
go test ./...
```
