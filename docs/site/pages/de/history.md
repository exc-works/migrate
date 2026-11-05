# Verlauf und Ausgangsstand

Der Verlauf enthält Version, Dateiname und Hash. Vergleichen Sie ihn mit dem tatsächlichen Schema.

```sh
migrate status
migrate status --output json
```

| Status | Bedeutung | Maßnahme |
| --- | --- | --- |
| `pending` | Nicht ausgeführt | Prüfen, dann `up` |
| `applied` | Erfolgreich ausgeführt | Datei unverändert lassen |
| `baseline` | Ohne Ausführung erfasst | Übernommenen Verlauf erhalten |
| `outOfOrder` | Ältere offene Version | Reihenfolge prüfen |
| `hashMismatch` | Inhalt stimmt nicht überein | Original wiederherstellen |
| `filenameMismatch` | Dateiname weicht ab | Originalnamen wiederherstellen |

Abweichungen blockieren. Neue Änderungen brauchen neue Versionen. Skripthashes umfassen keine [importierten Module](./modules#checksums).

## Bestehende Datenbank übernehmen {#baseline}

Nur verwenden, wenn Daten und Schema bereits den Dateien entsprechen:

1. Dateien mit dem Istzustand vergleichen.
2. Falls nötig `migrate create` ausführen.
3. `migrate baseline`, dann `migrate status` prüfen.
4. Weitere Änderungen mit höheren Versionen ergänzen.

Baseline erfasst ohne Ausführung und repariert keine teilweise fehlgeschlagene Migration.
