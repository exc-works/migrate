import { normalizeName } from "../shared/names.ts";

export function up(db: any) {
  for (const row of db.query("SELECT id, original_name FROM users")) {
    db.exec("UPDATE users SET display_name=? WHERE id=?", [
      normalizeName(row.original_name), row.id,
    ]);
  }
}

export function down(db: any) {
  db.exec("UPDATE users SET display_name=original_name");
}
