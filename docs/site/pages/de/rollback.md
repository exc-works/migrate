# Migrationen zurücknehmen

Die Rücknahme führt Down oder `down(db)` aus und kann nur die dort implementierte Umkehr leisten. Gelöschte Daten werden nicht automatisch wiederhergestellt.

## Zielversion behalten

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

Nur höhere Versionen werden zurückgenommen. Bei 1, 2, 3 entfernt `down 2` nur 3 und behält 1 sowie 2.

## Alle zurücknehmen

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

Ziel und `--all` schließen sich aus. Erfolgreich zurückgenommene Dateien werden wieder `pending`.

Den gesamten Bereich und die [Transaktionen](./script-transactions) prüfen. Ein fehlendes `down(db)` blockiert den Bereich; eine leere Funktion macht nichts rückgängig. DDL und nichttransaktionale Änderungen können einen Fehler überstehen. Vor Wiederherstellung Daten und Verlauf prüfen.
