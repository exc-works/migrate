import { normalizeName } from "../shared/names.ts";

export function up(db: any) {
  for (const row of db.query("SELECT id, original_name FROM users")) {
    db.exec("UPDATE users SET display_name=? WHERE id=?", [
      normalizeName(row.original_name), row.id,
    ]);
  }
  if (db.query("SELECT display_name FROM users WHERE id=1")[0].display_name !== "alice") {
    throw Error("Name normalization failed");
  }
}

export function down(db: any) {
  db.exec("UPDATE users SET display_name=original_name");
}
