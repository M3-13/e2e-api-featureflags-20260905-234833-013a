VERDICT: BUGS_FOUND

- **Title**: Go-Test-Suite schlägt fehl: `TestRoutesReachable` bewertet 404-Antworten registrierter Routen als fehlend
- **Symptom**: `go test ./...` läuft nicht grün (Exit-Code 1, `FAIL featureflags`). AC-12 („go test ./... läuft grün durch“) ist damit nicht erfüllt; die CI-Pipeline bricht ab.
- **Repro**: `go test ./...` ausführen. Der Fehler tritt in den Untertests `GET /flags/some-key` und `DELETE /flags/some-key` auf.
- **Evidence**:
  - `--- FAIL: TestRoutesReachable (0.00s)`
  - `--- FAIL: TestRoutesReachable/GET_/flags/some-key (0.00s)`
  - `routing_test.go:45: GET /flags/some-key answered 404, want it registered`
  - `--- FAIL: TestRoutesReachable/DELETE_/flags/some-key (0.00s)`
  - `routing_test.go:45: DELETE /flags/some-key answered 404, want it registered`
- **Suspected file(s)**: `routing_test.go` – die Testlogik erwartet offenbar, dass eine registrierte Route nicht mit 404 antwortet, obwohl `GET /flags/{key}` und `DELETE /flags/{key}` laut Spec bei unbekanntem Key korrekt 404 mit JSON-Fehlerobjekt liefern müssen (AC-05, AC-07). Die 404-Antwort ist hier das erwartete Produktverhalten, nicht das Fehlen der Route; der Test selbst ist zu streng/falsch und blockiert den grünen Gesamtlauf.
- **Severity**: high