import { normalizeName } from "../shared/names.ts";

export function up(db) {
  const originalName = " BOB ";
  db.exec("INSERT INTO users VALUES (?,?,?)", [
    2, originalName, normalizeName(originalName),
  ]);
  if (db.query("SELECT display_name FROM users WHERE id=2")[0].display_name !== "bob") {
    throw Error("Shared helper was not applied");
  }
}

export function down(db) {
  db.exec("DELETE FROM users WHERE id=?", [2]);
}
