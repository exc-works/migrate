# Migrationen ausführen

Prüfen Sie Zielverbindung und Dateien vor dem Deployment.

## Verlauf initialisieren

```sh
migrate create
```

Erzeugt bei Bedarf nur die Verlaufstabelle. Sind Änderungen bereits vorhanden, zuerst [Baseline](./history#baseline) prüfen.

## Vorschau und Ausführung

```sh
migrate up --dry-run
migrate up
migrate status
```

Offene Versionen werden geordnet ausgeführt und Erfolge gespeichert. Wiederholungen spielen erfolgreiche Versionen nicht erneut ab. Auch ohne Ausgabe kann die Ausführung erfolgreich sein; `status` prüfen.

Dry-run kann Verlauf lesen und SSH öffnen, führt aber keine Änderungen aus. Skripte werden geprüft, nicht ausgewertet; dynamisches SQL ist daher nicht vorhersehbar.

Ältere offene Versionen werden standardmäßig übersprungen. `migrate_out_of_order` nur nach Prüfung der Kompatibilität zum aktuellen Schema aktivieren.

## Deployment

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

Variablen und Geheimnisse vorab bereitstellen und Aufrufe auf dieselbe Datenbank serialisieren. Der Verlauf garantiert bei Parallelität, Abstürzen oder impliziten Commits keine genau einmalige Ausführung. Vor Wiederholungen den Zustand prüfen; siehe [Fehlerbehebung](./troubleshooting).
