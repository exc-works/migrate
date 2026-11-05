# Fehlerbehebung

Mit `migrate version`, `migrate capabilities --output json` und `migrate status --output json` beginnen. Nur status braucht eine DB-Verbindung.

## Befehl oder Datei fehlt

PATH unter [Installation](./installation) prüfen. Konfiguration im Projekt erzeugen oder `-c` / `-w` nutzen. YAML-Position, Werte und Einrückung prüfen: Leerzeichen statt Tabs. [Suchreihenfolge](./configuration#configuration-file-search-order) beachten.

## Variable fehlt

`can't find env value for XXX` bedeutet, dass `${XXX}` fehlt. Setzen oder für geeignete nichtgeheime Werte `${XXX:default}` verwenden. SSH-`password_env` enthält einen Variablennamen.

## Format, Datenbank oder Modul nicht unterstützt

`capabilities` prüfen und alte Versionen aktualisieren. Bei `unsupported dialect` die [Aliase](./databases#aliases) prüfen. Relative Pfade prüfen und npm vorab installieren. Eingebettete Quellen brauchen für npm `DependencyDir`. Node- oder asynchrone Abhängigkeiten bleiben ungeeignet.

## Verlauf weicht ab oder up ändert nichts

Originaldateien und Namen aus Git wiederherstellen, dann höhere Version hinzufügen. Nicht mit Baseline bei unbekanntem Zustand umgehen. Stille Befehle können erfolgreich sein; applied, baseline und outOfOrder prüfen.

## Rücknahme oder Verbindung fehlgeschlagen

`to-version must be set, or use --all`: Ziel oder `--all` wählen. Ziel bleibt erhalten. Skripte ohne down lassen sich nicht zurücknehmen.

Vor Wiederholung Daten und Verlauf untersuchen: DDL, nichttransaktionale Operationen und Abbrüche um COMMIT können Änderungen hinterlassen. Baseline repariert das nicht. SSH-Zugang, known_hosts, Proxy und Erreichbarkeit vom Bastion-Host prüfen. Keine automatische Wiederverbindung oder Wiederholung. Geschützte Serverlogs nutzen, ohne Zugangsdaten oder Zeilendaten zu teilen.
