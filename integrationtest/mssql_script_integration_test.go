//go:build integration && full

package integrationtest

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/script"
)

func runMSSQLScalarScenario(t *testing.T, db *sql.DB) {
	t.Helper()
	query := `SELECT CAST(1 AS BIT) AS flag,CAST('922337203685477.5807' AS MONEY) AS amount,CAST('-214748.3648' AS SMALLMONEY) AS small_amount,CAST('00112233-4455-6677-8899-aabbccddeeff' AS UNIQUEIDENTIFIER) AS guid,CONVERT(VARCHAR(36),CAST('00112233-4455-6677-8899-aabbccddeeff' AS UNIQUEIDENTIFIER)) AS guid_text,CAST('2026-11-03T01:02:03.1234567+08:00' AS DATETIMEOFFSET(7)) AS instant,CAST(N'中文' AS NVARCHAR(20)) AS label,CAST('<root/>' AS XML) AS xml,CAST(0x00FF AS VARBINARY(2)) AS payload`
	program, err := script.Compile("V0__mssql_scalars.js", `export function up(db){const r=db.query(`+strconv.Quote(query)+`)[0];if(r.flag!==true||r.amount!=='922337203685477.5807'||r.small_amount!=='-214748.3648'||r.guid.length!==16||r.guid_text.toLowerCase()!=='00112233-4455-6677-8899-aabbccddeeff'||r.instant!=='2026-11-02T17:02:03.1234567Z'||r.label!=='中文'||(r.xml!=='<root/>'&&r.xml!=='<root />')||r.payload.length!==2||r.payload[1]!==255)throw Error('MSSQL scalar fidelity')}`, "up", script.Options{})
	if err != nil {
		t.Fatal(err)
	}
	invocation, err := program.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = invocation.Execute(context.Background(), db, dialect.NameMSSQL); err != nil {
		t.Fatal(err)
	}
}
