//go:build full

package script

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMSSQLBatchesRejectedBeforeDriver(t *testing.T) {
	for _, query := range []string{
		"UPDATE users SET profile=@p1 COMMIT TRANSACTION",
		"SELECT 1 ROLLBACK TRANSACTION",
		"SELECT 1 COMMIT\u00a0TRANSACTION",
		"SELECT 1 /* outer /* nested */ text */ COMMIT",
		"SELECT 1 EXEC('COMMIT')",
		"SELECT 1 SET IMPLICIT_TRANSACTIONS ON",
		"UPDATE users SET profile=@p1 SET XACT_ABORT OFF",
		"SELECT 1 SELECT 2",
		"INSERT INTO users VALUES (@p1) DELETE FROM users",
		"SELECT 1 CREATE TABLE extra(id INT)",
		"WITH c AS (SELECT 1 AS id) SELECT id FROM c COMMIT",
		"ALTER TABLE users DROP COLUMN email DROP TABLE other_users",
		"ALTER TABLE users ALTER COLUMN email NVARCHAR(200) ALTER TABLE other_users ADD id INT",
		"ALTER TABLE users DROP CONSTRAINT constraint_name COMMIT TRANSACTION",
	} {
		for _, method := range []string{"exec", "query"} {
			t.Run(method+"/"+query, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				p, err := Compile("V1__batch.js", `export function up(db){try{db.`+method+`(`+strconv.Quote(query)+`,['changed'])}catch(e){}}`, "up", Options{})
				if err != nil {
					t.Fatal(err)
				}
				i, err := p.Instantiate(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				err = i.Execute(context.Background(), db, "sqlserver")
				if err == nil || !strings.Contains(err.Error(), "MSSQL") {
					t.Fatalf("batch reached driver or escaped failure: %v", err)
				}
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestMSSQLDirectStatementsAndQuoting(t *testing.T) {
	for _, query := range []string{
		"SELECT 'COMMIT', N'it''s ROLLBACK', [commit], [a]]rollback], \"EXEC\" FROM users",
		"SELECT @commit, [set] FROM users /* outer /* COMMIT */ ROLLBACK */ -- EXEC",
		"SELECT CASE WHEN id=1 THEN 'COMMIT' ELSE 'ROLLBACK' END FROM users",
		"SELECT id FROM users UNION ALL SELECT id FROM other_users",
		"WITH c AS (SELECT id FROM users) UPDATE users SET profile=@p1 WHERE id IN (SELECT id FROM c)",
		"INSERT INTO users(id) SELECT id FROM other_users UNION SELECT id FROM more_users",
		"CREATE VIEW current_users AS SELECT id FROM users",
		"DROP TABLE IF EXISTS users",
		"ALTER TABLE users DROP COLUMN email",
		"ALTER TABLE [users] ALTER COLUMN [email] NVARCHAR(200)",
		"ALTER TABLE users DROP CONSTRAINT [email_unique]",
		"ALTER TABLE users DROP /* nested /* comment */ here */ COLUMN email",
		"SELECT 中文 FROM 用户 WHERE 姓名=@p1",
	} {
		if got, err := prepareMSSQL(query); err != nil || got != query {
			t.Fatalf("%s: %v", query, err)
		}
	}
	merge := "WITH c AS (SELECT @p1 AS id) MERGE users AS t USING c ON t.id=c.id WHEN MATCHED THEN UPDATE SET profile=@p2 WHEN NOT MATCHED THEN INSERT(id,profile) VALUES(c.id,@p2) -- trailing comment"
	if got, err := prepareMSSQL(merge); err != nil || !strings.HasSuffix(got, "\n;") {
		t.Fatalf("MERGE terminator: %v", err)
	}
	for _, query := range []string{"SELECT 'unterminated", "SELECT [unterminated", "SELECT 1 /* unterminated", "SELECT (1"} {
		if _, err := prepareMSSQL(query); err == nil {
			t.Fatalf("accepted malformed query %s", query)
		}
	}
}
