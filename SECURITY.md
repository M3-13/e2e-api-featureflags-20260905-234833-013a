VERDICT: CHANGES_REQUESTED

## Sicherheitsreport

Prüfumfang: sämtliche sichtbaren Go-Quellen. Kein anwendbarer Scanner-Output vorhanden. Es sind keine externen Abhängigkeiten sichtbar; der Code nutzt ausschließlich die Standardbibliothek.

### Befund 1 — Mittel: Admin-API-Key wird nicht in konstanter Zeit verglichen

**Betroffene Stelle:** `auth.go`, in `withAdminAuth`

```go
if r.Header.Get("X-API-Key") != adminKey {
```

**Warum:** Der Vergleich `!=` bricht bei der ersten abweichenden Byte-Position ab. Dadurch hängt die Antwortzeit von der Übereinstimmungslänge ab. Ein Angreifer, der die API erreichen kann, kann den geheimen `ADMIN_API_KEY` byteweise über Timing-Messungen rekonstruieren. Da der Server aktuell nur an `127.0.0.1` lauscht, ist die praktische Ausnutzbarkeit begrenzt; sobald der Dienst über einen Reverse-Proxy exponiert wird, wird der Angriff realistisch.

**Konkreter Fix:** Geheimen Vergleich mit `crypto/subtle.ConstantTimeCompare` durchführen, idealerweise über einen Hash fester Länge, damit nicht einmal die Länge des Schlüssels durch Timing verraten wird.

```go
import (
    "crypto/sha256"
    "crypto/subtle"
)

func apiKeyMatches(actual, expected string) bool {
    a := sha256.Sum256([]byte(actual))
    b := sha256.Sum256([]byte(expected))
    return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}
```

Dann in `withAdminAuth`:

```go
if !apiKeyMatches(r.Header.Get("X-API-Key"), adminKey) {
    writeError(w, http.StatusUnauthorized, "unauthorized")
    return
}
```

---

### Befund 2 — Niedrig: Log-Injection über nicht bereinigten URL-Pfad

**Betroffene Stelle:** `middleware.go`

```go
logger.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
```

**Warum:** `r.URL.Path` ist bereits URL-dekodiert. Ein Angreifer kann Steuerzeichen wie `%0A` im Pfad senden; dadurch können unerwünschte Zeilenumbrüche oder ANSI-Steuerzeichen in Logdateien bzw. Terminal-Logs entstehen. Das kann Logzeilen verfälschen oder Angriffe verschleiern.

**Konkreter Fix:** Den Pfad mit `r.URL.EscapedPath()` oder `%q` ausgeben:

```go
logger.Printf("%s %s %d %s", r.Method, r.URL.EscapedPath(), sw.status, time.Since(start))
```

Alternativ:

```go
logger.Printf("%s %q %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
```

---

### Sonstige Prüfpunkte

- **Secrets:** Keine hartkodierten Produktionsschlüssel sichtbar. `ADMIN_API_KEY` wird aus der Umgebung gelesen. Testwerte wie `"secret"` in `_test.go` sind unkritisch.
- **Injection/Eingaben:** `POST`/`PUT` begrenzen den Body vor dem Einlesen auf 1 MiB (`http.MaxBytesReader`). Flag-Keys werden auf `[A-Za-z0-9._-]` validiert, `user` auf 1–128 Zeichen. JSON wird typsicher in Structs dekodiert. Keine SQL-, Command-, Path- oder SSRF-Injection erkennbar.
- **AuthN/AuthZ:** `POST`, `PUT`, `DELETE` sind über `X-API-Key` geschützt; bei nicht konfiguriertem Schlüssel antwortet der Server mit 503 statt die Endpunkte offen zu lassen. Lesende Endpunkte bleiben bewusst offen, was bei reinem Localhost-Betrieb akzeptabel ist. Vor einer Netzwerkexposition sollte geprüft werden, ob `GET /flags` und `GET /flags/{key}` zusätzlich geschützt werden müssen.
- **Dependencies:** Keine externen Pakete; nur Standardbibliothek. Keine ausnutzbaren CVEs sichtbar.
- **Transport/Konfiguration:** `X-Content-Type-Options: nosniff` wird bei JSON-Antworten gesetzt. 500er-Antworten werden korrekt auf `{"error":"internal server error"}` maskiert. Der Server bindet an `127.0.0.1` und setzt `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` sowie `MaxHeaderBytes`. Damit ist die Angriffsfläche derzeit begrenzt.