VERDICT: BUGS_FOUND

- **Titel**: Routing-Test wertet 404 bei unbekanntem Flag-Key fälschlich als fehlende Route
- **Symptom**: `go test ./...` bricht mit Exit-Code 1 ab. Die Subtests `GET /flags/some-key` und `DELETE /flags/some-key` schlagen fehl, obwohl der Server gemäß AC-05/AC-07 für unbekannte Keys korrekt 404 mit JSON-Fehlerobjekt liefert. Dadurch ist AC-12 („go test ./... läuft grün durch") verletzt.
- **Repro**: `go test ./...` im Projektverzeichnis ausführen.
- **Evidence**:
  ```
  --- FAIL: TestRoutesReachable (0.00s)
      --- FAIL: TestRoutesReachable/GET_/flags/some-key (0.00s)
          routing_test.go:45: GET /flags/some-key answered 404, want it registered
      --- FAIL: TestRoutesReachable/DELETE_/flags/some-key (0.00s)
          routing_test.go:45: DELETE /flags/some-key answered 404, want it registered
  ```
- **Suspected file(s)**: `routing_test.go` – der Test prüft offenbar auf einen Status ungleich 404 (oder ein anderes Kriterium), obwohl 404 bei unbekanntem Key laut Spezifikation korrekt ist. Der Handler in `flags_read.go` verhält sich spezifikationskonform.
- **Severity**: high