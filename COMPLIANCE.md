VERDICT: CHANGES_REQUESTED

## Strukturierter Compliance-Bericht

### 1. DSGVO (Datenschutz)

**Befund 1 — Pfad-Logging kann über Flag-Keys personenbezogene Daten verarbeiten**
- **Schweregrad:** hoch
- **Sachverhalt:** Die `LoggingMiddleware` in `middleware.go` protokolliert `r.URL.Path` im Klartext. Der Flag-Key ist Teil des Pfads, z. B. `GET /flags/alice@example.com/evaluate`. Die Key-Validierung (`^[A-Za-z0-9._-]+$`) lässt Zeichenfolgen zu, die personenbezogene Daten enthalten können (E-Mail-Adresse, Nutzerkennung). Enthält ein Flag-Key solche Daten, erscheinen sie unverschlüsselt im Log. Zusätzlich sind `GET /flags` und `GET /flags/{key}` ohne Authentifizierung abrufbar; auch `description` (bis 500 Zeichen) kann frei befüllt werden.
- **Risiko:** Unkontrollierte Verarbeitung personenbezogener Daten in Logs und öffentlich abrufbaren Antworten. Verstoß gegen Datenminimierung und ggf. Verarbeitung ohne dokumentierte Rechtsgrundlage.
- **Konkrete Abhilfe:**
  - In `COMPLIANCE.md`, `README.md` und `SECURITY.md` verbindlich dokumentieren: **„Flag-Keys und -Descriptions dürfen keine personenbezogenen Daten enthalten. Keys dienen ausschließlich als technische Bezeichner.“**
  - Alternativ/ergänzend die LoggingMiddleware anpassen (`middleware.go`): Den konkreten `{key}`-Segmentwert maskieren, z. B. mit einer Hilfsfunktion `maskPath(r.URL.Path)`, die das mittlere Segment durch `{key}` ersetzt, sodass geloggt wird: `GET /flags/{key}`, `PUT /flags/{key}`, `GET /flags/{key}/evaluate`. Damit bleibt der Log für Operationen nützlich, ohne potenzielle PII zu protokollieren.

**Befund 2 — Rechtsgrundlage und Verarbeitungszweck des `user`-Parameters nicht explizit dokumentiert**
- **Schweregrad:** mittel
- **Sachverhalt:** Der Endpunkt `GET /flags/{key}/evaluate?user=...` verarbeitet den Parameter `user` transient. Der Code speichert ihn nicht dauerhaft und protokolliert ihn nicht (Query-String wird weggelassen). Das ist datenschutzfreundlich. Es fehlt aber eine explizite Aussage zur Rechtsgrundlage und zum Zweck der Verarbeitung.
- **Risiko:** Unklare Verantwortlichkeit, wenn der Dienst produktiv eingesetzt wird. Der Betreiber kann nicht auf einen Blick erkennen, dass er eine Verarbeitung nachweisen muss.
- **Konkrete Abhilfe:** In `COMPLIANCE.md` einen Abschnitt **„Verarbeitung des Evaluate-user-Parameters“** aufnehmen: Zweck (deterministische Auslieferung von Feature-Flags), Rechtsgrundlage (z. B. Art. 6 Abs. 1 lit. f DSGVO bei berechtigtem Interesse, abhängig vom konkreten Einsatzszenario), Datenminimierung (nur für die Dauer der Anfrage verarbeitet, keine Speicherung, keine Weitergabe, kein Logging).

**Befund 3 — JSON-Body wird nur teilweise dekodiert; Rest wird verworfen**
- **Schweregrad:** niedrig
- **Sachverhalt:** In `flags_write.go` wird der Body mit `json.NewDecoder(r.Body).Decode(&req)` gelesen. Es wird nicht geprüft, ob nach dem ersten JSON-Objekt weitere Nicht-Whitespace-Zeichen folgen. Ein Body wie `{"key":"a","enabled":true}{"key":"b","enabled":true}` würde akzeptiert und der zweite Teil ignoriert.
- **Risiko:** Verdeckte Datenübermittlung bzw. unerwartetes Verhalten. Für die DSGVO ist das weniger direkt, aber Sicherheits- und Integritätsrisiko.
- **Konkrete Abhilfe:** In `flags_write.go` nach dem ersten `Decode` eine zweite Decodierung in eine leere Struktur durchführen und prüfen, dass `io.EOF` zurückkommt. Andernfalls 400 mit `{"error":"invalid JSON body"}`.

### 2. EU Cyber Resilience Act (CRA)

**Befund 4 — Kein dokumentierter Update-/Patch-Prozess und TLS-Vorgabe**
- **Schweregrad:** mittel
- **Sachverhalt:** Der Server bindet in `main.go` ausschließlich an `127.0.0.1` und spricht unverschlüsseltes HTTP. Das ist als sicherer Entwicklungs-Default gut. Für den produktiven Netzwerkbetrieb ist TLS erforderlich, aber weder im Code konfigurierbar noch in `SECURITY.md` explizit vorgeschrieben. Auch die Art und Weise, wie Updates des Dienstes eingespielt werden, ist nicht dokumentiert.
- **Risiko:** Im Produktiveinsatz unzureichende Absicherung der Transportstrecke; unklare Update-Verantwortlichkeiten.
- **Konkrete Abhilfe:** In `SECURITY.md` einen Abschnitt **„Betrieb und Updates“** ergänzen: Der Dienst ist in Produktion hinter einem TLS-terminierenden Reverse-Proxy zu betreiben; Updates erfolgen durch Austausch des Binaries/Containers; Stillstandzeiten sind unkritisch, da keine Persistenz besteht. Optional zusätzlich eine Konfigurationsmöglichkeit für TLS-Zertifikate im Code bereitstellen.

**Befund 5 — Kein Rate-Limiting auf offenen Endpunkten**
- **Schweregrad:** mittel
- **Sachverhalt:** Die offenen Endpunkte (`GET /healthz`, `GET /flags`, `GET /flags/{key}`, `GET /flags/{key}/evaluate`) sind ungeschützt durch `withAdminAuth`. Ein Angreifer mit Netzwerkzugriff könnte den Dienst durch viele Anfragen beeinträchtigen. Ein Rate-Limiting fehlt.
- **Risiko:** Verfügbarkeitsrisiko, das dem Grundsatz „Security by default“ des CRA widerspricht.
- **Konkrete Abhilfe:** In `SECURITY.md` festhalten, dass vorgelagert ein Rate-Limiter eingesetzt werden soll; optional im Code eine einfache Begrenzung (z. B. Token-Bucket) für die offenen Endpunkte implementieren. Die Implementierung darf die eigenen Tests und die Funktion nicht beeinträchtigen — Tests nutzen wenige Requests, daher unbedenklich.

**Befund 6 — SBOM-Prozess nicht verankert**
- **Schweregrad:** niedrig
- **Sachverhalt:** Die Datei `sbom.spdx.json` ist vorhanden. Da keine externen Abhängigkeiten verwendet werden (nur Standardbibliothek), ist das SBOM überschaubar. Es ist jedoch nicht geregelt, wann es aktualisiert werden muss (z. B. bei künftigen `go.mod`-Änderungen).
- **Risiko:** Veraltete SBOM, dadurch möglicher Verstoß gegen CRA-Transparenzpflichten.
- **Konkrete Abhilfe:** In `SECURITY.md` oder `COMPLIANCE.md` einen Prozessbaustein aufnehmen: **„Bei jeder Änderung von go.mod oder Hinzufügen von Abhängigkeiten ist `sbom.spdx.json` unmittelbar zu aktualisieren.“**

### 3. EU AI Act

- **Prüfung:** Im Produkt ist keine KI-Funktion implementiert. Es handelt sich um eine reine Konfigurations-/Auslieferungs-API. Der AI Act findet keine Anwendung.

### 4. Pflichttexte & UI

- **Prüfung:** Das Produkt ist ein reines Backend ohne Endnutzer-UI. Es bestehen keine Pflichten zu Impressum, Cookie-Banner, Widerrufsbelehrung oder vergleichbaren Kunden-UI-Texten.
- **Hinweis:** Für den Betreiber des Dienstes kann dennoch eine externe Datenschutzerklärung erforderlich sein, die beschreibt, dass der `user`-Parameter transient verarbeitet wird. Das fällt jedoch nicht in den Code dieses Repositories und ist als Betreiberpflicht dokumentiert (siehe Befund 2).

### 5. Barrierefreiheit (WCAG/BITV/EAA)

- **Prüfung:** Da keine öffentliche Web-UI vorhanden ist, bestehen keine Barrierefreiheitsanforderungen. Die JSON-Antworten sind maschinenlesbar und benötigen keine WCAG-Konformität.

---

### Zusammenfassung der positiven Aspekte

- AC-19 wird nachweislich erfüllt: Die LoggingMiddleware nutzt `r.URL.Path` und lässt den Query-String (`user`) aus; entsprechende Tests sind vorhanden.
- AC-14: Body-Limit von 1 MiB ist korrekt umgesetzt.
- AC-15/AC-16: Key- und User-Validierung entspricht den Vorgaben.
- AC-17/AC-18: `X-Content-Type-Options: nosniff` und Maskierung von 500-Fehlern sind implementiert und getestet.
- Der In-Memory-Store speichert keine personenbezogenen Daten dauerhaft; der `user`-Wert wird nicht persistiert.
- Der Standard-Bind an `127.0.0.1` ist ein sicherer Default.
- `sbom.spdx.json` ist vorhanden.

**Gesamtbewertung:** Es liegen keine fundamentalen Verstöße vor, die eine Sperrung erfordern. Die genannten Punkte sind behebbar und betreffen vor allem Dokumentation, Log-Maskierung und optionale Härtung. Daher `CHANGES_REQUESTED`.