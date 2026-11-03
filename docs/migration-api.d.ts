/** Editor declarations only. The full CLI provides the implementation. */
type MigrationSQLParameter =
  | string
  | number
  | boolean
  | null
  | readonly number[]
  | Uint8Array;

type MigrationSQLValue = string | number | boolean | null | number[];

/** Ordinary SQL errors can be caught for recovery. Cancellation, connection loss,
 * parameter/result validation and resource-limit failures remain fatal.
 */
interface MigrationDB {
  readonly dialect: "postgres" | "mysql" | "mariadb" | "sqlite" | "tidb" | "redshift" | "oracle" | "clickhouse" | "mssql";

  /** Integer/decimal results are strings. Results are materialized and bounded. */
  query(
    sql: string,
    params?: readonly MigrationSQLParameter[],
  ): Record<string, MigrationSQLValue>[];

  /** Use native placeholders. Counts are driver-reported decimal strings. Native SQL is forwarded unchanged; use config.transaction=false for manual transactions. */
  exec(
    sql: string,
    params?: readonly MigrationSQLParameter[],
  ): { rowsAffected: string };
}

/** Optional module export: export const config: MigrationConfig = { transaction: false }.
 * Applies to up/down. Defaults to true; false has no automatic rollback or data/history atomicity.
 */
interface MigrationConfig {
  readonly transaction?: boolean;
}

/** Synchronous return values are ignored. Promise results are unsupported. */
type MigrationEntrypoint = (db: MigrationDB) => unknown;
