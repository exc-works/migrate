# Go से उपयोग

root पैकेज का सार्वजनिक API SemVer का पालन करता है; internal/* सार्वजनिक नहीं। कॉलर pool बंद करता, उपयुक्त context देता और समान DB के माइग्रेशन क्रमिक चलाता है।

Config.DryRun और DryRunOutput योजना दिखाते हैं, लेकिन Create() इतिहास बनाता है। डिफ़ॉल्ट logger NoopLogger है। [स्क्रिप्ट एम्बेड](./modules) में FSSource और निर्भरताएँ दें।

OpenConnection के साथ Connection.Context() से Ping और सेवा चलाएँ, अंत में Close() करें। नेटवर्क ड्राइवर पंजीकृत हैं, SQLite कॉलर पंजीकृत करे। ConnectionOptions.Prompt और HostKeyPrompt वैकल्पिक हैं और cancel मानें। callbacks बिना इनपुट या अनजान host पर विश्वास नहीं होता।

CLI के अलावा, `github.com/exc-works/migrate` को आपके सर्विस कोड से सीधे import करके माइग्रेशन चलाए जा सकते हैं — यह यूनिट टेस्ट, स्टार्टअप हुक या एडमिन पैनल के लिए उपयोगी है।

## इंस्टॉलेशन

```bash
go get github.com/exc-works/migrate
```

अपनी आवश्यकता का डेटाबेस ड्राइवर import करें (library किसी एक से बंधी नहीं है):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

## न्यूनतम उदाहरण

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // idempotent: यदि history table न हो तो बनाता है
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

## मुख्य API

- `migrate.NewService(ctx, migrate.Config)` माइग्रेशन executor बनाता है
- `svc.Create()` history table `migration_schema` बनाता है (idempotent)
- `svc.Up()` सभी pending माइग्रेशन लागू करता है
- `svc.Down(toVersion, all)` लक्ष्य वर्शन या सब कुछ rollback करता है
- `svc.Status()` `[]migrate.MigrationStatus` लौटाता है
- `svc.Baseline()` मौजूदा pending फ़ाइलों को `baseline` के रूप में चिह्नित करता है

सामान्य प्रकार:

- Dialects (कंस्ट्रक्टर को प्राथमिकता दें — ये `Dialect` interface लौटाते हैं): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, या नाम से lookup के लिए `migrate.DialectFromName("postgres")`
- Sources: `DirectorySource` (फ़ाइल सिस्टम), `StringSource` (in-memory slice), `FSSource` (कोई भी `fs.FS`, जैसे `//go:embed` या `os.DirFS`), `CombinedSource` (कई sources को मिलाता है)
- Loggers: `migrate.NoopLogger{}` (डिफ़ॉल्ट), `migrate.NewStdLogger("info", os.Stdout)`, या कोई भी `migrate.Logger` interface लागू करने वाला प्रकार

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
