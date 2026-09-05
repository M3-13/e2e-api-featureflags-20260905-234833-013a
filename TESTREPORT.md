VERDICT: BUGS_FOUND

- **Title**: TestRoutesReachable erwartet fälschlich kein 404 für unbekannten Flag-Key und blockiert `go test`
- **Symptom**: `go test ./...` schlägt fehl, obwohl die Anwendung laut Spezifikation korrekt arbeitet: AC-05 und AC-07 verlangen 404 für unbekannte Flags. Dadurch ist die CI rot und AC-12 („go test ./... läuft grün durch“) nicht erfüllt.
- **Repro**: `go test ./...` ausführen; der Test `TestRoutesReachable` ruft `GET /flags/some-key` und `DELETE /flags/some-key` auf, ohne den Key `some-key` vorher im Store anzulegen.
- **Evidence**:
  ```
  --- FAIL: TestRoutesReachable (0.00s)
      --- FAIL: TestRoutesReachable/GET_/flags/some-key (0.00s)
          routing_test.go:45: GET /flags/some-key answered 404, want it registered
      --- FAIL: TestRoutesReachable/DELETE_/flags/some-key (0.00s)
          routing_test.go:45: DELETE /flags/some-key answered 404, want it registered
  ```
- **Suspected file(s)**: `routing_test.go` – insbesondere die Funktion `TestRoutesReachable`, die die Route mit einem unbekannten Key prüft und fälschlich einen Status ungleich 404 erwartet. Die Produkthandler `handleGet` und `handleDelete` in `flags_read.go` verhalten sich spezifikationsgemäß; der Fehler liegt ausschließlich im Test.
- **Severity**: high