/** Editor declarations only. The full CLI provides the implementation. */
type MigrationSQLParameter =
  | string
  | number
  | boolean
  | null
  | readonly number[]
  | Uint8Array;

type MigrationSQLValue = string | number | boolean | null | number[];

interface MigrationDB {
  readonly dialect: "postgres" | "mysql" | "mariadb" | "sqlite" | "tidb" | "redshift";

  /** Integer/decimal results are strings. Results are materialized and bounded. */
  query(
    sql: string,
    params?: readonly MigrationSQLParameter[],
  ): Record<string, MigrationSQLValue>[];

  /** Use native placeholders. Transaction/session control is unsupported. */
  exec(
    sql: string,
    params?: readonly MigrationSQLParameter[],
  ): { rowsAffected: string };
}
